# internal/config/ — per-machine settings

See the inventory row in `internal/README.md` for the public API.

## What lives here

- **`config.go`**: `Config` (role, name, box, rc_dir), validation, TOML load/save at `~/.config/roost/config.toml` (honours `XDG_CONFIG_HOME`).

## Naming

`<area>.go` with `<area>_test.go` beside it; fixtures in `testdata/`.

## Read first if

You're adding a setting.

## Skip / don't read

Test files, unless you're changing behaviour.

## Discipline

- A new key needs: a field with a toml tag, validation, a flag in `cli init`, and a README mention.
- Names must be valid hostnames (they become tailnet names).

## Cross-references

- **Consumes:** `host`.
- **Feeds:** `cli`, `step`, `steps`.
