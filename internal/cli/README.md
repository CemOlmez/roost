# internal/cli/ — cobra commands

See the inventory row in `internal/README.md` for the public API.

## What lives here

- **`root.go`**: `App` (dependencies), `Execute` (exit codes), root command, helpers.
- **`setup.go`**: `init`, `doctor`, `confirm` and the summary output.
- **`ops.go`**: `status`, `attach`, `rc`, `version`.
- **`cli_test.go`**: end-to-end command tests against `hosttest.Fake`.

## Naming

`<area>.go`; command tests all live in `cli_test.go`.

## Read first if

You're adding or changing a command or flag. Also read `.claude/skills/golang-cli/SKILL.md`.

## Skip / don't read

Test files, unless you're changing behaviour.

## Discipline

- Write to `app.Out`/`app.Err`, never `os.Stdout`. Return errors and never call `os.Exit`.
- "Something needs attention" returns `errUnhealthy` (exit 1, no extra message).
- Commands stay thin: setup behaviour belongs in a step, not in a command.

## Cross-references

- **Consumes:** `step`, `steps`, `config`, `host`, `tailscale`, `ui`.
- **Feeds:** `cmd/roost`.
