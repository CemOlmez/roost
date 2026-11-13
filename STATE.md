# STATE — Roost

Current snapshot. **Rewrite lines in place when state changes — never append history.**
The journal is `NOTES.md`; git owns commit history.

> **▶ Picking this up fresh?** Read `tasks/handoff.md` first.

## As of 2026-11-13

- **Phase:** v0.1.0 built, audited (correctness, security, docs) and tested; release pending tag; not yet validated on real box hardware
- **Stack:** Go 1.27, cobra, BurntSushi/toml; GoReleaser; GitHub Actions. Linux only (amd64, arm64, armv7).

## What exists (built + verified)

- `roost init` / `doctor` / `status` / `attach` / `rc` / `confirm` / `version`, plus shell completion
- 13 setup steps (see `internal/steps/README.md`), every one unit-tested with `hosttest.Fake`; a `-tags live` test for the Windows bridge
- `doctor` and `status` verified on the author's laptop (WSL, client role, and box role via temp config)
- `install.sh` verified end-to-end against a local release build, including checksum-tamper refusal
- CI: tests (race), golangci-lint, shellcheck, goreleaser snapshot, govulncheck, CodeQL, Dependabot; actions pinned by SHA; release has provenance attestation
- Pre-release audit findings fixed (see NOTES 2026-11-13)

## What's next

- First real box setup on the home PC (`tasks/backlog.md` MVP); log results in `NOTES.md`

## Known gaps / risks

- Unverified on real hardware: RoostKeepAlive keeping WSL alive after reboot, `claude remote-control` running headless under systemd+tmux (first-run prompts?), the elevated script under a real UAC prompt.
- `tailscale up` inside WSL vs Tailscale on Windows: only the WSL variant is supported and tested.
