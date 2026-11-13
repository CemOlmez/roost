# internal/steps/assets/ — files embedded in the binary

Compiled into roost with `go:embed` and read via `steps.Asset(name)`.

## What lives here

- **`roost-rc.service.tmpl`**: systemd user unit for Claude Remote Control (Go `text/template`: Path, ConfigDir, Dir, Tmux, Claude, Name). Uses its own tmux socket (`-L roost-rc`) and `Restart=always`. Check compares the rendered output with the installed file, so editing this marks existing installs "out of date" and `init` rewrites them.
- **`roost-windows.ps1`**: runs once as administrator, passed in memory via `-EncodedCommand` with `$Distro` and `$TaskUser` prepended: RoostKeepAlive task (hidden window, runs as the invoking user), power settings, hibernate off, wake-on-magic-packet. Prints to its own window and waits for Enter.

## Naming

`roost-<what>.<ext>`.

## Read first if

You're changing the Remote Control service or Windows-side setup.

## Skip / don't read

—

## Discipline

- The PowerShell script must stay idempotent (`-Force`, set-not-toggle), must never ask for or store a password, and must not write files to user-writable places (it runs elevated).
- Keep the check in `windows.go` (`windowsHostCheck`) in sync with what the script sets.

## Cross-references

- **Consumes:** —
- **Feeds:** `steps/claude.go` (unit), `steps/windows.go` (script).
