# cmd/ — binaries

Entry points only. Logic lives in `internal/`, never here.

## What lives here

- **`roost/`**: the `roost` binary. See `roost/README.md`.

## Naming

One folder per binary, named after the binary.

## Read first if

You're changing how the binary starts (signals, version injection).

## Skip / don't read

Nothing here is large; for anything else go to `internal/`.

## Discipline

- `main` only wires dependencies and calls `cli.Execute`. No flags, no logic.

## Cross-references

- **Consumes:** `internal/cli`, `internal/host`.
- **Feeds:** `.goreleaser.yaml` (`main: ./cmd/roost`, ldflags set `main.version`/`main.commit`).
