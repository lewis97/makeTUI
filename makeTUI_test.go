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
