# internal/steps/assets/ — files embedded in the binary

Compiled into roost with `go:embed` and read via `steps.Asset(name)`.

## What lives here

- **`roost-rc.service.tmpl`**: systemd user unit for Claude Remote Control (Go `text/template`: Path, Dir, Tmux, Claude, Name). Check compares the rendered output with the installed file, so editing this marks existing installs "out of date" and `init` rewrites them.
- **`roost-windows.ps1`**: runs once as administrator: RoostKeepAlive task, power settings, hibernate off, wake-on-magic-packet. Writes `roost-windows.log` next to itself, which init prints.

## Naming

`roost-<what>.<ext>`.

## Read first if

You're changing the Remote Control service or Windows-side setup.

## Skip / don't read

—

## Discipline

- The PowerShell script must stay idempotent (`-Force`, set-not-toggle) and must never ask for or store a password.
- Keep the check in `windows.go` (`windowsHostCheck`) in sync with what the script sets.

## Cross-references

- **Consumes:** —
- **Feeds:** `steps/claude.go` (unit), `steps/windows.go` (script).
