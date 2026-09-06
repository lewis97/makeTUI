# makeTUI

`makeTUI` is a terminal picker for Makefile targets. It lets you search the
available targets, inspect their description or recipe, and run one without
leaving the terminal.

## Requirements

- Go 1.24 or later
- GNU Make (or a compatible `make` command)

## Build and run

Build the executable:

```sh
make build
```

Then run it from a directory containing a `Makefile`:

```sh
./dist/maketui
```

The program exits before it runs the selected `make <target>` command, so the
command output appears in your regular terminal.

## Controls

| Key | Action |
| --- | --- |
| Up / Down | Move through targets |
| Type | Fuzzy-search target names |
| Backspace | Remove the last search character |
| Ctrl+U | Clear the search |
| Enter | Run the selected target |
| Esc / Ctrl+C | Quit |

## Makefile descriptions

Add a `##` comment immediately before a target to show a friendly description:

```make
## Build the application
build:
	go build ./...
```

If a target has no `##` comment, makeTUI displays its recipe in the detail
panel instead:

```make
test:
	go test ./...
```

Only ordinary target declarations are shown; special targets such as `.PHONY`
are omitted.
