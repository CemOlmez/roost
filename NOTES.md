# Iteration notes

Log of real setups: what was run, what worked, what broke. Each entry feeds the next fix.

## Iteration 1 — 2026-11-13, laptop (HP, WSL Ubuntu 24.04, client)

`roost doctor` before any changes:
- ✓ systemd, packages, Tailscale installed, Claude installed + logged in, phone on tailnet (offline for 47 days)
- ✗ Tailscale SSH off. Fix: `roost init --role client` (needs sudo).

Observed:
- Claude sessions were running directly in terminals, not tmux, so they die when the lid closes. This is exactly what `roost attach` on the box solves.
- Tailscale runs inside WSL (not on Windows). Keep it that way on the box too, so Tailscale SSH lands directly in Linux.

## Iteration 2 — home PC (box) — TODO

To verify on real hardware:
- [ ] Does the RoostKeepAlive task (`wsl.exe ... sleep infinity`) keep the distro and systemd services alive after logon, with no window open?
- [ ] Does `vmIdleTimeout=-1` have any effect alongside the task, or is the task enough?
- [ ] Does `roost-rc` (tmux + `claude remote-control`) show up in the Claude app after a full Windows reboot?
- [ ] Does `claude remote-control` need anything interactive on first run (trust dialog, workspace prompt)? If so, first run should go through `roost rc attach`.
- [ ] BIOS menu names on the real motherboard.
- [ ] Is RC_DIR=$HOME the right default for spawned sessions, or should it be the workspace folder?
