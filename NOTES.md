# NOTES — Roost

Running journal. Newest first. One dated entry per meaningful session: what shipped,
decisions and why, deviations from plan, real-hardware test logs. `STATE.md` says what is
true; this says how it got that way.

---

## 2026-11-13 — v0.1.0: rewrite as a Go CLI, repo published

**Shipped:** Go rewrite of the Bash prototype: `cmd/roost`, `internal/{cli,step,steps,host,config,tailscale,ini,ui}`, 13 steps, unit tests on fakes, CI/security/release workflows, GoReleaser, checksum-verifying `install.sh`, docs, project structure (CLAUDE.md, folder READMEs, trackers).

**Decisions:**
- **Go single binary** over Bash/Python/C#. See `docs/decisions/26-11-13_go-single-binary.md`.
- **cobra, no viper.** Config is four keys in one TOML file; viper's layering isn't worth the dependency.
- **Steps as data (struct with Check/Fix/Manual funcs)**, not an interface per step: less boilerplate, the whole checklist reads top to bottom in `steps.All()`.
- **Tailscale SSH instead of OpenSSH server.** No sshd to harden or expose; auth is the tailnet identity. `--operator=$USER` so roost can read and change Tailscale without sudo afterwards.
- **Remote Control runs as `claude remote-control` (server mode) inside tmux under a systemd user service.** tmux gives it a TTY and lets `roost rc attach` show the QR code; systemd restarts it and starts it on boot (linger).
- **WSL keep-alive = Windows scheduled task at logon running `wsl.exe -d <distro> --exec sleep infinity`**, plus `.wslconfig vmIdleTimeout=-1`. Unverified; see the test log below.
- **Windows checks read the registry and hex powercfg values only**, because the author's Windows display language isn't English.
- **BIOS can't be detected**, so the user runs `roost confirm bios`.

**Prototype lesson (Bash version, same day):** `roost doctor` on the laptop showed Claude sessions running directly in terminals, not tmux, so they die with the lid. That's the core problem roost exists for.

### Test log — laptop (HP, WSL Ubuntu 24.04), 2026-11-13

- doctor as client: 6/7 ✓. ✗ tailnet name is "CemOlmez" (want "laptop"), Tailscale SSH off. Expected; init fixes it.
- doctor as box: Windows checks parse correctly (hibernate on, no task, STANDBY_AC=0), motherboard detected via WMI.

### Test checklist — home PC (box), TODO

- [ ] `roost init` from a clean `curl | sh` install
- [ ] UAC prompt appears; `roost-windows.ps1` log shows OK lines
- [ ] after a Windows reboot (no window opened): `devbox` shows in the Claude app; `roost status` on the laptop shows it online
- [ ] does `claude remote-control` need an interactive first run (trust prompt) before the service works? If yes → make `roost rc attach` part of init
- [ ] pull power → PC boots itself (BIOS) → auto-login → box back online
- [ ] laptop: `roost attach` lands in the box's tmux; VS Code Remote-SSH to `devbox` works with Tailscale SSH
