# Handoff — Roost

Standing orientation for a fresh session. **Rewritten in place when stale.**

## Read in this order

1. `CLAUDE.md` (automatic)
2. `STATE.md`, in full
3. `NOTES.md`, top 2 entries
4. `tasks/backlog.md`, `## MVP` only
5. The folder README for today's work

## Tracker protocol — one event, one place

When a task completes: `NOTES.md` entry → extract deferrals into `backlog.md` → delete
the completed backlog item → rewrite the changed lines in `STATE.md`.

## Locked decisions — do not re-litigate

- Go single binary, cobra, no viper. Settled 2026-11-13; see `docs/decisions/26-11-13_go-single-binary.md`.
- No Roost server/account/telemetry. Settled 2026-11-13: trust and simplicity are the product.
- Tailscale SSH, not OpenSSH. Settled 2026-11-13: nothing to expose or harden.
- Linux-only binary; Windows is reached from WSL via `powershell.exe`. Settled 2026-11-13.

## Constraints & gotchas

- The author's Windows display language isn't English. Never parse localized command output.
- `sudo` needs a password on the author's machines. Agents can run `doctor`/`status`, not `init`.
- `tailscale status --json` exits non-zero when logged out but still prints JSON (handled in `tailscale.GetStatus`).
- Testing `install.sh` locally: set `ROOST_BIN_DIR` *and* make sure it's on PATH, or it appends a PATH line to the real `~/.bashrc`.

## How to verify

```bash
make ci
```
