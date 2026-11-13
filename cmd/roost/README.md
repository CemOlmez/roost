# cmd/roost/ — the roost binary

`main.go` builds a `cli.App` with the real `host.Local{}`, stdio and version info, runs it with a signal-aware context, and exits with its code.

## What lives here

- **`main.go`**: the only file. `version`/`commit` are set at release time via `-ldflags -X`.

## Naming

`main.go` only.

## Read first if

You're changing version injection or exit handling.

## Skip / don't read

—

## Discipline

- Keep it under ~40 lines. Anything testable belongs in `internal/cli`.

## Cross-references

- **Consumes:** `internal/cli` (`App`, `Execute`), `internal/host` (`Local`).
- **Feeds:** release builds (`.goreleaser.yaml`, `Makefile` LDFLAGS).
