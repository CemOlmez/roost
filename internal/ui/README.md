# internal/ui/ — terminal output and prompts

See the inventory row in `internal/README.md` for the public API.

## What lives here

- **`ui.go`**: `UI` with `OK` ✓, `Fail` ✗, `Warn` !, `Skip`, `Hint`, `Heading`, `Confirm`, `Ask`. Colors only on a TTY and without `NO_COLOR`; `Yes` answers defaults (for `--yes`).

## Naming

`<area>.go` with `<area>_test.go` beside it; fixtures in `testdata/`.

## Read first if

You change how roost looks or asks.

## Skip / don't read

Test files, unless you're changing behaviour.

## Discipline

- Steps print through `env.UI`, never fmt.Print.
- Detect terminals with `golang.org/x/term`, not `ModeCharDevice`: `/dev/null` is a character device too (that bug let `init </dev/null` run).

## Cross-references

- **Consumes:** —
- **Feeds:** `step`, `cli`.
