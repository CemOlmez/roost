# internal/host/ — the only door to the OS

See the inventory row in `internal/README.md` for the public API.

## What lives here

- **`host.go`**: `Host` interface + `Local` (real implementation).
- **`platform.go`**: WSL/systemd detection, PowerShell bridge, Windows path helpers.
- **`hosttest/`**: the fake. See `hosttest/README.md`.

## Naming

`<area>.go` with `<area>_test.go` beside it; fixtures in `testdata/`.

## Read first if

You need a new kind of OS interaction. Add it to `Host` and `hosttest.Fake` together.

## Skip / don't read

Test files, unless you're changing behaviour.

## Discipline

- Imports nothing from roost.
- `Run` returns trimmed stdout with CRs removed (Windows output); errors include stderr.
- `RunTTY` for anything that may prompt (sudo, installers, UAC launch).

## Cross-references

- **Consumes:** —
- **Feeds:** every other package.
