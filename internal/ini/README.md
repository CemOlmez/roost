# internal/ini/ — WSL config file editing

See the inventory row in `internal/README.md` for the public API.

## What lives here

- **`ini.go`**: `Get`/`Set` that keep comments, order and unknown keys (case-insensitive like WSL).

## Naming

`<area>.go` with `<area>_test.go` beside it; fixtures in `testdata/`.

## Read first if

You touch `/etc/wsl.conf` or `.wslconfig` handling.

## Skip / don't read

Test files, unless you're changing behaviour.

## Discipline

- Never rewrite a whole user file. Only `Set` the one key.

## Cross-references

- **Consumes:** —
- **Feeds:** `steps` (wsl-systemd, wslconfig).
