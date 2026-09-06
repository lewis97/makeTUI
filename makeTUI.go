package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type target struct {
	name   string
	desc   string
	recipe string
}

type model struct {
	targets  []target
	filtered []target
	cursor   int
	query    string
	width    int
	height   int
	selected string
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

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true)
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

	m := model{
		targets:  targets,
		filtered: targets,
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "make-tui: %v\n", err)
		os.Exit(1)
	}

	// The TUI must be fully restored before handing terminal control to make.
	if final, ok := finalModel.(model); ok && final.selected != "" {
		if err := runTarget(final.selected); err != nil {
			fmt.Fprintf(os.Stderr, "make-tui: %v\n", err)
			os.Exit(1)
		}
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "down":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "enter":
			if len(m.filtered) == 0 {
				return m, nil
			}

			m.selected = m.filtered[m.cursor].name
			return m, tea.Quit

		case "backspace":
			if len(m.query) > 0 {
				// Remove the last rune rather than the last byte.
				runes := []rune(m.query)
				m.query = string(runes[:len(runes)-1])
				m.applyFilter()
			}

		case "ctrl+u":
			m.query = ""
			m.applyFilter()

		default:
			// Anything printable is treated as search input.
			if msg.Type == tea.KeyRunes {
				m.query += string(msg.Runes)
				m.applyFilter()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	}

	return m, nil
}

func (m *model) applyFilter() {
	m.filtered = fuzzyFilter(m.targets, m.query)
	m.cursor = 0
}

func runTarget(name string) error {
	cmd := exec.Command("make", name)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m model) View() string {
	if len(m.targets) == 0 {
		return ""
	}

	var b strings.Builder

	// Header.
	b.WriteString(titleStyle.Render(" MAKE "))
	b.WriteString(" ")
	b.WriteString(dimStyle.Render("targets"))
	b.WriteString("\n\n")

	// Search input.
	b.WriteString(searchStyle.Render("/ "))
	b.WriteString(m.query)

	if m.query == "" {
		b.WriteString(dimStyle.Render("type to search..."))
	}

	b.WriteString("\n\n")

	// Target list.
	listHeight := m.height - 11
	if listHeight < 3 {
		listHeight = 3
	}

	if len(m.filtered) == 0 {
		b.WriteString(dimStyle.Render("  No matching targets"))
		b.WriteString("\n")
	} else {
		start := 0
		end := len(m.filtered)

		if len(m.filtered) > listHeight {
			start = m.cursor - listHeight/2

			if start < 0 {
				start = 0
			}

			end = start + listHeight

			if end > len(m.filtered) {
				end = len(m.filtered)
				start = end - listHeight
			}
		}

		for i := start; i < end; i++ {
			t := m.filtered[i]

			line := "  " + t.name

			if i == m.cursor {
				line = "▶ " + t.name

				b.WriteString(
					selectedStyle.Render(
						padRight(line, m.width-4),
					),
				)
			} else {
				b.WriteString(normalStyle.Render(line))
			}

			b.WriteString("\n")
		}
	}

	// Description box.
	b.WriteString("\n")

	if len(m.filtered) > 0 {
		selected := m.filtered[m.cursor]

		description := selected.desc
		if description == "" {
			description = selected.recipe
		}

		if description == "" {
			description = dimStyle.Render("No recipe defined.")
		}

		descWidth := m.width - 4
		if descWidth < 20 {
			descWidth = 20
		}

		desc := descTitleStyle.Render(selected.name) +
			"\n" +
			descStyle.Render(description)

		b.WriteString(
			borderStyle.
				Width(descWidth).
				Render(desc),
		)
	} else {
		b.WriteString(
			borderStyle.
				Width(max(20, m.width-4)).
				Render(dimStyle.Render("No matching target")),
		)
	}

	b.WriteString("\n\n")

	// Footer.
	b.WriteString(
		dimStyle.Render(
			"↑/↓ navigate   type search   ctrl+u clear   enter run   esc quit",
		),
	)

	return b.String()
}

func padRight(s string, width int) string {
	if width <= len(s) {
		return s
	}

	return s + strings.Repeat(" ", width-len(s))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// fuzzyFilter performs a lightweight fuzzy match.
//
// Characters from the query must occur in order in the target name,
// but don't need to be consecutive.
//
// For example:
//
//	query "bt" matches "build-test"
//	query "ga" matches "generate-assets"
//	query "cl" matches "clean"
//
// Results are scored so tighter matches and matches near the beginning
// of the target are preferred.
func fuzzyFilter(targets []target, query string) []target {
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		return targets
	}

	type result struct {
		target target
		score  int
	}

	var results []result

	for _, t := range targets {
		score, ok := fuzzyScore(strings.ToLower(t.name), query)

		if ok {
			results = append(results, result{
				target: t,
				score:  score,
			})
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	filtered := make([]target, 0, len(results))

	for _, r := range results {
		filtered = append(filtered, r.target)
	}

	return filtered
}

func fuzzyScore(text, query string) (int, bool) {
	if query == "" {
		return 0, true
	}

	textRunes := []rune(text)
	queryRunes := []rune(query)

	queryIndex := 0
	score := 0

	lastMatch := -1

	for i, r := range textRunes {
		if queryIndex >= len(queryRunes) {
			break
		}

		if r != queryRunes[queryIndex] {
			continue
		}

		// Reward matches near the beginning.
		score += 10

		if i == 0 {
			score += 20
		}

		// Reward consecutive characters.
		if lastMatch == i-1 {
			score += 15
		}

		// Reward matches after separators.
		if i > 0 {
			switch textRunes[i-1] {
			case '-', '_', '/', '.', ' ':
				score += 12
			}
		}

		// Penalize gaps between matches.
		if lastMatch >= 0 {
			score -= i - lastMatch - 1
		}

		lastMatch = i
		queryIndex++
	}

	if queryIndex != len(queryRunes) {
		return 0, false
	}

	// Exact match is strongly preferred.
	if text == query {
		score += 1000
	}

	// Prefix match is preferred.
	if strings.HasPrefix(text, query) {
		score += 100
	}

	return score, true
}

// parseMakefile extracts targets, their preceding ## comments, and recipes.
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
func parseMakefile(filename string) ([]target, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	targetRe := regexp.MustCompile(`^([A-Za-z0-9_./-]+)\s*:(?:[^=]|$)`)

	var targets []target
	var pendingDesc string
	var current *target

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Documentation comment.
		if strings.HasPrefix(trimmed, "##") {
			pendingDesc = strings.TrimSpace(
				strings.TrimPrefix(trimmed, "##"),
			)
			current = nil
			continue
		}

		// Recipe lines belong to the most recently declared target. Strip the
		// leading tab so the detail panel shows the shell code clearly.
		if strings.HasPrefix(line, "\t") {
			if current != nil {
				if current.recipe != "" {
					current.recipe += "\n"
				}
				current.recipe += strings.TrimPrefix(line, "\t")
			}
			continue
		}

		// Ignore blank lines while retaining a pending description and target.
		if trimmed == "" {
			continue
		}

		// Ignore non-recipe indented lines.
		if line[0] == ' ' {
			continue
		}

		match := targetRe.FindStringSubmatch(line)
		if match == nil {
			pendingDesc = ""
			current = nil
			continue
		}

		name := match[1]

		// Ignore special/internal targets.
		if strings.HasPrefix(name, ".") {
			pendingDesc = ""
			current = nil
			continue
		}

		targets = append(targets, target{
			name: name,
			desc: pendingDesc,
		})
		current = &targets[len(targets)-1]

		pendingDesc = ""
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}
