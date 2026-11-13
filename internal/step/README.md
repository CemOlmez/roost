# internal/step/ — the check/fix engine

See the inventory row in `internal/README.md` for the public API.

## What lives here

- **`step.go`**: `Step`, `Env`, `Result`, `Status`, `ErrRestartRequired`.
- **`run.go`**: `Run` (ModeCheck/ModeFix), `Report`, `Outcome`.

## Naming

`<area>.go` with `<area>_test.go` beside it; fixtures in `testdata/`.

## Read first if

You're changing how steps are run, reported or confirmed. Adding a *step* is `internal/steps/`.

## Skip / don't read

Test files, unless you're changing behaviour.

## Discipline

- Knows nothing about specific steps. If you need a step's name here, the design is wrong.
- Every behaviour change needs a case in `run_test.go`.

## Cross-references

- **Consumes:** `config`, `host`, `ui`.
- **Feeds:** `steps`, `cli`.
