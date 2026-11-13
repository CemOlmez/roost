# NOTES — Roost

Running journal. Newest first. One dated entry per meaningful session: what shipped,
decisions and why, deviations from plan, real-hardware test logs. `STATE.md` says what is
true; this says how it got that way.

---

## 2026-11-13 — Pre-release audit: correctness, security, docs

Three independent reviews before going public. Fixed:

**Correctness**
- Remote Control shared tmux's default socket with the user's terminals. That could kill the service right after start, and `rc restart` could kill the user's `main` terminal (same cgroup). → dedicated socket `tmux -L roost-rc`.
- `Restart=on-failure` never fires (the tmux server exits 0 when Claude dies), and `is-active` could pass with Claude gone. → `Restart=always` + `tmux has-session` in Check; dropped the meaningless `After=network-online.target` in a user unit.
- Unit PATH missed nvm/npm installs; `CLAUDE_CONFIG_DIR` wasn't passed. → claude's own dir first in PATH; env passed through.
- `tailscale up --flags` refuses when the node has other non-default prefs. → `tailscale set`, then a bare `up` only if down.
- A duplicate tailnet name (`devbox-1`) passed the check. → compare the MagicDNS label.
- Keep-alive task opened a console window (closing it killed the keep-alive), and ran as the UAC admin, not the user. → hidden PowerShell window, explicit principal = invoking user.
- Invalid config → doctor said "All set. 0 checks passed." → `Load` validates; an empty report is unhealthy.
- `roost confirm phone` "worked" but nothing read it. → `Confirmable` flag.
- Mixed UTF-8/UTF-16 log from Windows PowerShell 5.1. → no log file at all (see below).

**Security**
- The elevated script sat in `%TEMP%` while UAC waited, so it was swappable. → passed in memory via `-EncodedCommand`, inputs bound as escaped PS literals, distro name validated. Verified live (`go test -tags live`): apostrophes survive, 7.5k chars encoded.
- SECURITY.md claimed tailnet-only access; Remote Control is a second path via the Claude account. → documented both paths, recommended SSH policy, kill switch.
- `--operator=$USER` let any user process (incl. a prompt-injected agent) reconfigure Tailscale. → removed; reads don't need it.
- `rc_dir`/paths flowed unescaped into the unit. → `config.SafePath` everywhere.
- Fixed-name staging file before `sudo install`. → `CreateTemp` + `install -T -o root -g root` + remove.
- Supply chain: actions pinned by SHA, `persist-credentials: false`, exact GoReleaser version, no `go mod tidy` hook at release, tests in a read-only job, build-provenance attestation.
- install.sh: HTTPS-only curl, BusyBox-compatible checksum, real TTY test, never writes a custom `ROOST_BIN_DIR` into shell rc files.

**Docs:** first-run expectations, WSL basics, PATH reopen note, VS Code caveat, same-username rule, complete "what changes" + undo, update/uninstall, DESIGN marked planned vs built.

**Deferred → backlog:** `box_user`, tailnet SSH-policy warning, `rc_dir` default.

---

## 2026-11-13 — v0.1.0: rewrite as a Go CLI, repo published

**Shipped:** Go rewrite of the Bash prototype: `cmd/roost`, `internal/{cli,step,steps,host,config,tailscale,ini,ui}`, 13 steps, unit tests on fakes, CI/security/release workflows, GoReleaser, checksum-verifying `install.sh`, docs, project structure (CLAUDE.md, folder READMEs, trackers).

**Decisions:**
- **Go single binary** over Bash/Python/C#. See `docs/decisions/26-11-13_go-single-binary.md`.
- **cobra, no viper.** Config is four keys in one TOML file; viper's layering isn't worth the dependency.
- **Steps as data (struct with Check/Fix/Manual funcs)**, not an interface per step: less boilerplate, the whole checklist reads top to bottom in `steps.All()`.
- **Tailscale SSH instead of OpenSSH server.** No sshd to harden or expose; auth is the tailnet identity. (An `--operator` grant was later removed in the audit.)
- **Remote Control runs as `claude remote-control` (server mode) inside tmux under a systemd user service.** tmux gives it a TTY and lets `roost rc attach` show the QR code; systemd restarts it and starts it on boot (linger).
- **WSL keep-alive = Windows scheduled task at logon running `wsl.exe -d <distro> --exec sleep infinity`**, plus `.wslconfig vmIdleTimeout=-1`. Unverified; see the test log below.
- **Windows checks read the registry and hex powercfg values only**, because the author's Windows display language isn't English.
- **BIOS can't be detected**, so the user runs `roost confirm bios`.

**Prototype lesson (Bash version, same day):** `roost doctor` on the laptop showed Claude sessions running directly in terminals, not tmux, so they die with the lid. That's the core problem roost exists for.

### Test log — laptop (HP, WSL Ubuntu 24.04), 2026-11-13

- doctor as client: 6/7 ✓. ✗ tailnet name is the old device name (want "laptop"), Tailscale SSH off. Expected; init fixes it.
- doctor as box: Windows checks parse correctly (hibernate on, no task, STANDBY_AC=0), motherboard detected via WMI.

### Test checklist — home PC (box), TODO

- [ ] `roost init` from a clean `curl | sh` install
- [ ] UAC prompt appears; `roost-windows.ps1` log shows OK lines
- [ ] after a Windows reboot (no window opened): `devbox` shows in the Claude app; `roost status` on the laptop shows it online
- [ ] does `claude remote-control` need an interactive first run (trust prompt) before the service works? If yes → make `roost rc attach` part of init
- [ ] pull power → PC boots itself (BIOS) → auto-login → box back online
- [ ] laptop: `roost attach` lands in the box's tmux; VS Code Remote-SSH to `devbox` works with Tailscale SSH
