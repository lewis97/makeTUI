package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMakefileCapturesRecipes(t *testing.T) {
	makefile := `## Build the application
build:
	go build ./...

clean:
	rm -rf dist
	@echo cleaned
`

	path := filepath.Join(t.TempDir(), "Makefile")
	if err := os.WriteFile(path, []byte(makefile), 0o600); err != nil {
		t.Fatal(err)
	}

	targets, err := parseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(targets) != 2 {
		t.Fatalf("got %d targets, want 2", len(targets))
	}

	if got, want := targets[0].desc, "Build the application"; got != want {
		t.Errorf("build description = %q, want %q", got, want)
	}
	if got, want := targets[0].recipe, "go build ./..."; got != want {
		t.Errorf("build recipe = %q, want %q", got, want)
	}
	if got, want := targets[1].recipe, "rm -rf dist\n@echo cleaned"; got != want {
		t.Errorf("clean recipe = %q, want %q", got, want)
	}
}

func TestParseJSONfile(t *testing.T) {
	dir := t.TempDir()
	config := `{"commands":[{"name":"test echo","description":"Print a test message","code":"echo test"}]}`
	if err := os.WriteFile(filepath.Join(dir, jsonConfigFileName), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(workingDir) })

	targets, err := parseJSONfile()
	if err != nil {
		t.Fatal(err)
	}

	if len(targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(targets))
	}
	if got, want := targets[0], (target{name: "test echo", desc: "Print a test message", recipe: "echo test", ttype: customTarget}); got != want {
		t.Errorf("target = %#v, want %#v", got, want)
	}
}

func TestCommandForTarget(t *testing.T) {
	tests := []struct {
		name   string
		target target
		want   []string
	}{
		{
			name:   "make target",
			target: target{name: "build", ttype: makeTarget},
			want:   []string{"make", "build"},
		},
		{
			name:   "custom target",
			target: target{name: "say hello", recipe: `echo "hello"`, ttype: customTarget},
			want:   []string{"sh", "-c", `echo "hello"`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd, err := commandForTarget(test.target)
			if err != nil {
				t.Fatal(err)
			}
			if got := cmd.Args; !equalStrings(got, test.want) {
				t.Errorf("command arguments = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestCommandForTargetRejectsEmptyCustomCommand(t *testing.T) {
	_, err := commandForTarget(target{name: "empty", ttype: customTarget})
	if err == nil {
		t.Fatal("expected an error for an empty custom command")
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
