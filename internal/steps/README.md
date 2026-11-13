# internal/steps/ — the setup checklist

Everything `roost init` sets up and `roost doctor` checks. The engine that runs these is `internal/step/`; this folder is the *what*. How to add one: `docs/ARCHITECTURE.md` § Adding a step.

## What lives here

- **`steps.go`**: `All()` (run order), `ByID`, `Confirm`, embedded `Asset`, shared helpers (`sudo`, `sudoWriteFile`, `shell`, `readText`).
- **`linux.go`**: WSL systemd, packages, linger. **`network.go`**: Tailscale, phone. **`claude.go`**: Claude Code + Remote Control service. **`windows.go`**: everything on the Windows host.
- **`assets/`**: embedded files. See `assets/README.md`.

## Inventory

Run order = table order. B = box, C = client. **Fix**: auto = roost fixes it; manual = the user acts.

| ID | Ensures | Roles | WSL only | Fix | How it checks |
|---|---|---|---|---|---|
| `wsl-systemd` | systemd is PID 1 in WSL | B C | ✓ | auto (sudo; needs `wsl --shutdown`) | `/proc/1/comm`, `/etc/wsl.conf` |
| `packages` | tmux, git, curl | B C | | auto (apt) | `LookPath` |
| `tailscale-installed` | tailscale CLI | B C | | auto (official script) | `LookPath` |
| `tailscale-connected` | logged in, hostname = name, SSH on, operator = user | B C | | auto (`tailscale up/set`) | `status --json`, `debug prefs` |
| `claude-installed` | Claude Code CLI | B C | | auto (official script) | `LookPath`, `~/.local/bin/claude` |
| `claude-login` | subscription login | B C | | manual (`/login`) | `$CLAUDE_CONFIG_DIR/.credentials.json` |
| `linger` | user services run without a session | B | | auto (`loginctl`) | `loginctl show-user` |
| `remote-control` | `roost-rc` user service running, unit up to date | B | | auto (unit + systemctl) | rendered template == file, `is-active` |
| `wslconfig` | `.wslconfig` `vmIdleTimeout=-1` | B | ✓ | auto (backup + edit) | `ini.Get` on Windows profile |
| `windows-host` | RoostKeepAlive task, no AC sleep, hibernate off, WoL | B | ✓ | auto (UAC, `roost-windows.ps1`) | registry + hex powercfg |
| `windows-autologon` | Windows logs in by itself | B | ✓ | manual (Sysinternals Autologon) | Winlogon `AutoAdminLogon` |
| `bios` | power on after AC loss, WoL | B | ✓ | manual (`roost confirm bios`) | marker file |
| `phone` | a phone on the tailnet | B C | | manual (apps) | tailnet peer OS android/iOS |

## Naming

Constructor = PascalCase of what it ensures (`TailscaleConnected`). ID = kebab-case, **stable forever** (users type IDs in `roost confirm`).

## Read first if

You're adding, changing or debugging a setup step.

## Skip / don't read

`steps_test.go` unless changing behaviour. `assets/roost-windows.ps1` unless touching Windows.

## Discipline

- Check is read-only. Fix is idempotent and ends where Check passes.
- Every new step: a test with `hosttest.Fake`, an inventory row here, and a row in README "What Roost changes".
- Order matters: a step may assume every earlier step passed (e.g. `remote-control` needs Claude installed).
- Windows: parse values, never localized words.

## Cross-references

- **Consumes:** `step` (types), `host`, `config`, `tailscale`, `ini`.
- **Feeds:** `cli` (`init`, `doctor`, `confirm`). Documented for users in root `README.md`.
