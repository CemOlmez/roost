# internal/tailscale/ — Tailscale state

See the inventory row in `internal/README.md` for the public API.

## What lives here

- **`tailscale.go`**: `GetStatus` (tolerates non-zero exit when logged out), `GetPrefs`, `Node.IsPhone`.
- **`testdata/`**: recorded `status --json` output.

## Naming

`<area>.go` with `<area>_test.go` beside it; fixtures in `testdata/`.

## Read first if

You need more Tailscale information. Add fields to the structs; record real output into `testdata/`.

## Skip / don't read

Test files, unless you're changing behaviour.

## Discipline

- Read-only. Changing Tailscale settings is a step's Fix (via sudo), not this package.

## Cross-references

- **Consumes:** `host`.
- **Feeds:** `steps`, `cli status`.
