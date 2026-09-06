# makeTUI

`makeTUI` is a terminal picker for Makefile targets and custom shell commands.
It lets you search the available commands, inspect their description or code,
and run one without leaving the terminal.

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

The program exits before it runs the selected command, so its output appears in
your regular terminal.

### zshrc update

To run via an alias:

```
alias mk="<PATH TO DIST BINARY>"
```

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

## Custom commands

Optionally add a `.maketui.json` file alongside the `Makefile` to include
commands that are not Make targets:

```json
{
  "commands": [
    {
      "name": "test echo",
      "description": "Print a test message",
      "code": "echo \"test echo\""
    }
  ]
}
```

Each entry requires a display `name` and the shell `code` to run. The optional
`description` is shown in the detail panel; when it is omitted, the `code` is
shown instead.

Makefile targets run as `make <target>`. Custom commands run exactly from their
`code` field using `sh -c`, so shell features such as pipes and variable
expansion are supported. Keep this file trusted, since its commands are
executed with your user account.
