# Changelog

All notable user-visible changes. Format: [Keep a Changelog](https://keepachangelog.com), versions: [SemVer](https://semver.org).

## [Unreleased]

## [0.1.0] - 2026-11-13

### Added
- `roost init`: interactive setup of a **box** (always-on dev machine) or **client** (laptop), safe to re-run.
- `roost doctor` (`--json`): read-only health check, exits 1 when something needs attention.
- `roost status` (`--json`), `roost attach`, `roost rc attach|logs|start|stop|restart`, `roost confirm bios`.
- Setup steps: WSL systemd, base packages, Tailscale (SSH, hostname, operator), Claude Code install + login check, systemd linger, Claude Remote Control as a user service in tmux, Windows `.wslconfig` idle timeout, Windows keep-alive task + power + wake-on-LAN (one UAC prompt), auto-login and BIOS guidance, phone check.
- `install.sh`: downloads the release for linux amd64/arm64/armv7 over HTTPS and verifies its checksum.

### Security
- The elevated Windows setup script is passed to PowerShell in memory (`-EncodedCommand`); nothing is written to `%TEMP%`.
- Remote Control runs in its own tmux server, isolated from your terminals.
- Root-owned files are staged in exclusively created temp files and installed as root:root.
- Paths written into the systemd unit are validated; Tailscale no longer grants `--operator`.
- Releases: actions pinned to commit SHAs, credentials not persisted, build-provenance attestations.
