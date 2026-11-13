# Architecture

How roost is built, and how to extend it. For *why* roost exists and where it's going, see [DESIGN.md](DESIGN.md).

## In one paragraph

roost is a single Go binary. Its core is a **checklist of steps**. Each step knows how to *check* one thing about the machine ("Tailscale SSH is on") and, usually, how to *fix* it. `roost doctor` runs every check; `roost init` runs every check and offers each fix. Every interaction with the real machine goes through one interface, `host.Host`, so every step is unit-tested against an in-memory fake.

## Packages

```
cmd/roost/            main: builds the App with the real host and runs the CLI
internal/
  cli/                cobra commands → step engine / host
  step/               the engine: Step type, Run(), Report
  steps/              the checklist itself (one constructor per step) + embedded assets
    assets/           systemd unit template, Windows PowerShell script (go:embed)
  host/               the only package that touches the OS (exec, files, env)
    hosttest/         in-memory fake Host for tests
  config/             ~/.config/roost/config.toml (role, name, box, rc_dir)
  tailscale/          parses `tailscale status --json` / `tailscale debug prefs`
  ini/                comment-preserving edits to wsl.conf / .wslconfig
  ui/                 checklist output (✓ ✗ !) and prompts
```

Dependency direction, no cycles:

```
cmd/roost → cli → steps → step → {host, config, ui}
                   ↘ tailscale, ini, host
```

`step` knows nothing about specific steps. `steps` knows nothing about cobra. `host` knows nothing about roost.

## The step model

```go
type Step struct {
    ID      string                                  // stable, kebab-case: "tailscale-connected"
    Title   func(*Env) string                       // the desired state, phrased as done
    Roles   []config.Role                           // box, client
    WSLOnly bool                                    // skipped on plain Linux / VPS / Pi
    Check   func(context.Context, *Env) Result      // must not change anything
    Fix     func(context.Context, *Env) error       // nil → manual step
    Manual  func(context.Context, *Env) []string    // instructions when the user must act
}
```

`step.Run(ctx, env, steps.All(), mode)` walks the list in order:

| Check result | Has Fix? | `doctor` (ModeCheck) | `init` (ModeFix) |
|---|---|---|---|
| pass | — | ✓ | ✓ |
| fail | yes | ✗ + detail | ✗, ask "Fix this now?", fix, re-check |
| fail | no | ! + instructions | ! + instructions |

A fix that only takes effect after a restart (enabling systemd in WSL) returns an error wrapping `step.ErrRestartRequired`. The run stops there and tells the user what to do.

### Rules every step follows

1. **Check is read-only and idempotent.** doctor must be safe to run any time, by anyone.
2. **Fix is idempotent.** Running init twice must be harmless.
3. **Fix ends in a state Check accepts.** The engine re-checks after every fix and reports "still not right" otherwise.
4. **Never handle the user's passwords.** sudo prompts go to the terminal (`RunTTY`); Windows auto-login is a manual step.
5. **Root-owned files** are staged in `~/.config/roost/.staging-*` and moved into place with `sudo install` (see `sudoWriteFile`).
6. **Windows checks parse values, not words.** Windows output is localized, so checks read registry values and hex values only.

## Adding a step

1. Write a constructor in the matching `internal/steps/*.go` file (`linux.go`, `network.go`, `claude.go`, `windows.go`), or a new file for a new area.
2. Add it to `All()` in `steps.go`, in the order it must run (dependencies first).
3. Add a row to the inventory in `internal/steps/README.md`.
4. Test it in `steps_test.go` with `hosttest.Fake`: script the commands it runs, assert on `Check` before and after `Fix`.
5. If it changes the machine, add it to "What Roost changes" in the root `README.md`.

`TestCatalogIsWellFormed` fails if a step lacks an ID, Title, Check or Roles, has neither Fix nor Manual, or reuses an ID.

## Testing

- `make test` runs the unit tests with `-race -shuffle=on`. No test touches the real machine.
- `hosttest.Fake` matches commands by their full command line. An unscripted command fails the test, so tests also document exactly what roost runs.
- Things only real hardware can show (a Windows reboot, a BIOS setting, a power cut) are tracked as manual test checklists in `NOTES.md` and fed back as fixes.

## Releases

Pushing a `v*` tag runs `.github/workflows/release.yml`: tests, then GoReleaser builds `linux/amd64`, `linux/arm64` and `linux/armv7` and attaches `roost_linux_<arch>.tar.gz` and `checksums.txt` to a GitHub Release. `install.sh` downloads from `releases/latest/download/` and verifies the checksum before installing.
