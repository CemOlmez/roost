# internal/host/hosttest/ — fake Host for tests

An in-memory `host.Host`. Files are a map, executables are a set, and commands are matched by their **full command line**; an unscripted command returns an error, so tests document exactly what roost runs.

## What lives here

- **`fake.go`**: `New()` (user `dev`, home `/home/dev`), `.On(cmdline, out, err)`, `.File(path, content)`, `.Dir(path)`, `.Bin(names...)`, `.Ran(cmdline)`, `.Calls`, `.Written` (last content per path, kept after `Remove`). `CreateTemp` names files `pattern` with `*` → `tmp`.

## Naming

—

## Read first if

You're writing a test for a step or command.

## Skip / don't read

—

## Discipline

- Test-only: never import from non-test code.
- When `Host` grows a method, the fake grows it in the same change.

## Cross-references

- **Consumes:** `internal/host`.
- **Feeds:** tests in `cli`, `config`, `step`, `steps`, `tailscale`.
