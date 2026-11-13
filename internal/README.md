# internal/ — all of roost's logic

Private packages (Go's `internal/` keeps them out of other modules). Dependency direction: `cli → steps → step → {host, config, ui}`; `steps` also uses `tailscale` and `ini`, and `cli` uses `config`, `host`, `tailscale` and `ui` directly. No cycles, and `host` imports nothing from roost.

## What lives here

Each package has its own README.

## Inventory

| Package | Responsibility | Public API | Depends on |
|---|---|---|---|
| `cli` | cobra commands; turns flags/config into step runs and output | `App`, `Execute`, `NewRootCmd` | step, steps, config, host, tailscale, ui |
| `step` | the engine: step type, run in check/fix mode, report | `Step`, `Env`, `Result`, `Pass`, `Failf`, `Run`, `Report`, `ErrRestartRequired` | config, host, ui |
| `steps` | the setup checklist and its embedded assets | `All`, `ByID`, `Confirm`, `Asset`, one constructor per step | step, host, config, tailscale, ini |
| `host` | the only OS access: commands, files, env; WSL/Windows helpers | `Host`, `Local`, `IsWSL`, `HasSystemd`, `PowerShell`, `WindowsEnvDir`, `WindowsPath`, `WSLToWindows`, `WSLDistro`, `IsLaptop`, `IsNotExist`; `Host` also has `CreateTemp`, `Remove` | — |
| `host/hosttest` | in-memory scripted `Host` for tests | `New`, `Fake.On/File/Dir/Bin/Ran`, `Fake.Written` | host |
| `config` | per-machine config file (role, name, box, rc_dir) | `Config`, `Role`, `Load` (validates), `Save`, `SafePath`, `Dir`, `Path`, `ErrNotFound` | host |
| `tailscale` | parse `tailscale status --json` / `debug prefs` | `GetStatus`, `ParseStatus`, `GetPrefs`, `Status`, `Node` (`MagicDNSName`, `IsPhone`), `Prefs` | host |
| `ini` | comment-preserving get/set for wsl.conf and .wslconfig | `Get`, `Set` | — |
| `ui` | checklist output (✓ ✗ ! -) and prompts | `UI`, `New`, `OK/Fail/Warn/Skip/Hint/Heading`, `Println/Printf`, `Bold/Dim`, `Confirm`, `Ask`, `IsNonTerminalFile` | — |

## Naming

Package = directory name, lowercase, singular-ish (`step` = the engine, `steps` = the catalog).

## Read first if

You're adding code anywhere: check the inventory above before creating a package or helper.

## Skip / don't read

`host/hosttest` unless you're writing tests.

## Discipline

- No `os/exec`, `os.ReadFile`, `os.Getenv` outside `host` (`cli` writes to `App.Out`, not `os.Stdout`).
- New package → a row here + its own README in the same change.

## Cross-references

- **Consumes:** —
- **Feeds:** `cmd/roost`. Described by `docs/ARCHITECTURE.md`.
