package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type target struct {
	name string
	desc string
}

type model struct {
	targets  []target
	cursor   int
	width    int
	height   int
	err      error
	running  bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("8")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("4")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	descTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)
)

func main() {
	targets, err := parseMakefile("Makefile")
	if err != nil {
		fmt.Fprintf(os.Stderr, "make-tui: %v\n", err)
		os.Exit(1)
	}

	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "make-tui: no targets found in Makefile")
		os.Exit(1)
	}

	p := tea.NewProgram(
		model{targets: targets},
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "make-tui: %v\n", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.targets)-1 {
				m.cursor++
			}

		case "enter":
			return m, runTarget(m.targets[m.cursor].name)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case commandFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, tea.Quit
	}

	return m, nil
}

type commandFinishedMsg struct {
	err error
}

func runTarget(name string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("make", name)

		// Restore the normal terminal temporarily so the command can
		// behave like a normal interactive command.
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()

		return commandFinishedMsg{err: err}
	}
}

func (m model) View() string {
	if len(m.targets) == 0 {
		return ""
	}

	var b strings.Builder

	// Header.
	header := titleStyle.Render(" MAKE ")
	b.WriteString(header)
	b.WriteString(" ")
	b.WriteString(dimStyle.Render("targets"))
	b.WriteString("\n\n")

	// Target list.
	listHeight := m.height - 8
	if listHeight < 3 {
		listHeight = 3
	}

	start := 0
	end := len(m.targets)

	if len(m.targets) > listHeight {
		start = m.cursor - listHeight/2

		if start < 0 {
			start = 0
		}

		end = start + listHeight

		if end > len(m.targets) {
			end = len(m.targets)
			start = end - listHeight
		}
	}

	for i := start; i < end; i++ {
		t := m.targets[i]

		line := "  " + t.name

		if i == m.cursor {
			line = "▶ " + t.name
			b.WriteString(selectedStyle.Render(padRight(line, m.width-4)))
		} else {
			b.WriteString(normalStyle.Render(line))
		}

		b.WriteString("\n")
	}

	// Description box.
	var description string

	if m.targets[m.cursor].desc != "" {
		description = m.targets[m.cursor].desc
	} else {
		description = dimStyle.Render("No description available.")
	}

	descWidth := m.width - 4
	if descWidth < 20 {
		descWidth = 20
	}

	desc := descTitleStyle.Render(m.targets[m.cursor].name) +
		"\n" +
		descStyle.Render(description)

	b.WriteString("\n")
	b.WriteString(borderStyle.Width(descWidth).Render(desc))
	b.WriteString("\n\n")

	// Footer.
	b.WriteString(
		dimStyle.Render("↑/k up   ↓/j down   enter run   q/esc quit"),
	)

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")).
				Render("command failed: " + m.err.Error()),
		)
	}

	return b.String()
}

func padRight(s string, width int) string {
	if width <= len(s) {
		return s
	}

	return s + strings.Repeat(" ", width-len(s))
}

// parseMakefile extracts targets and their preceding ## comments.
//
// Example:
//
//	## Build the application
//	build:
//	    go build ./...
//
//	## Run tests
//	test:
//	    go test ./...
//
// Only targets with a normal Makefile target declaration are included.
// Internal/private targets beginning with "." are ignored.
func parseMakefile(filename string) ([]target, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Match common Makefile targets while avoiding assignments and
	// indented recipe lines.
	targetRe := regexp.MustCompile(`^([A-Za-z0-9_./-]+)\s*:(?:[^=]|$)`)

	var targets []target
	var pendingDesc string

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		trimmed := strings.TrimSpace(line)

		// Capture comments immediately preceding a target.
		if strings.HasPrefix(trimmed, "##") {
			pendingDesc = strings.TrimSpace(strings.TrimPrefix(trimmed, "##"))
			continue
		}

		// Ignore blank lines while retaining a pending description.
		if trimmed == "" {
			continue
		}

		// Recipe/indented lines cannot be targets.
		if len(line) > 0 && (line[0] == '\t' || line[0] == ' ') {
			continue
		}

		match := targetRe.FindStringSubmatch(line)
		if match == nil {
			pendingDesc = ""
			continue
		}

		name := match[1]

		// Ignore special/internal targets.
		if strings.HasPrefix(name, ".") {
			pendingDesc = ""
			continue
		}

		targets = append(targets, target{
			name: name,
			desc: pendingDesc,
		})

		pendingDesc = ""
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

