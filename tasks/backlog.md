# Backlog — Roost

Undone work only. **Every item sits in exactly one tier**, and the tier is a judgement
about what the product *is*, not about effort or urgency.

| Tier | Test |
|---|---|
| **MVP** | Does v1's core loop break without it? *(install → init → close laptop → keep working from phone → box survives reboot)* |
| **Core** | Is it part of roost's identity: an always-reachable, never-lost dev box? |
| **Later** | Is it an improvement rather than a change of nature? |

**Item shape:** what · why · **trigger** · **pointer** · **ripple**.
**Capture is immediate.** **On completion:** NOTES entry → extract deferrals → delete the item.

---

## MVP — v1 is not done without these

- [ ] **[box · 26-11-13] Validate on a real Windows home PC.** Run the home-PC checklist in `NOTES.md`. **Trigger:** author at home with the PC. **Pointer:** `NOTES.md` § Test checklist. **Ripple:** every failure becomes a fix + test in `internal/steps/`.
- [ ] **[rc · 26-11-13] Confirm `claude remote-control` runs headless.** It may need a first interactive run (workspace trust) before working under systemd. **Trigger:** first box setup. **Pointer:** `internal/steps/claude.go` RemoteControlService. **Ripple:** maybe a new manual step "approve Remote Control once via `roost rc attach`".

## Core — the product's identity, not ship-blocking

- [ ] **[wake · 26-11-13] `roost wake` / `roost sleep` + Pi waker role.** Turn the box on from the phone via wake-on-LAN through an always-on device. **Trigger:** v0.1 validated. **Pointer:** `docs/DESIGN.md` § Waker. **Ripple:** new role `waker`, MAC address in config (captured from `roost-windows.ps1` log), new steps, README roadmap.
- [ ] **[sync · 26-11-13] Laptop live copy (Syncthing) + `roost here`.** Primary lease, conflict handling. **Trigger:** wake shipped. **Pointer:** `docs/DESIGN.md` § Concepts. **Ripple:** config gains peers; status shows sync state; big-workspace excludes.
- [ ] **[vault · 26-11-13] Encrypted off-site vault (restic) + `roost restore` / `leave --wipe`.** Covers the "everything is down, friend's laptop" scenario. **Trigger:** sync shipped. **Pointer:** `docs/DESIGN.md` § Scenario 3. **Ripple:** secrets handling, bucket setup docs.

## Later — refinements, ideas, and things that are not problems yet

- [ ] **[ux · 26-11-13] Undo command (`roost uninstall`).** README lists manual undo steps; automate them. **Trigger:** user request. **Pointer:** README § What Roost changes. **Ripple:** each step gains an optional Undo func.
- [ ] **[ux · 26-11-13] macOS client.** Laptop role on a Mac (no WSL steps apply). **Trigger:** a Mac user asks. **Pointer:** `.goreleaser.yaml` goos, `install.sh` OS check. **Ripple:** darwin builds; tmux/tailscale install differs.
- [ ] **[dash · 26-11-13] Web dashboard + browser terminal over the tailnet.** **Trigger:** v0.4. **Pointer:** `docs/DESIGN.md`. **Ripple:** new `roost serve`, auth model.
- [ ] **[docs · 26-11-13] Per-motherboard BIOS menu paths.** Collect from users; show in the BIOS step. **Trigger:** 3+ reports. **Pointer:** `internal/steps/windows.go` BIOS. **Ripple:** a data table.
