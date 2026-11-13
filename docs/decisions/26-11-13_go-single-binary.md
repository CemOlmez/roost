# Go single binary, cobra without viper

**Date:** 2026-11-13 · **Status:** accepted

## Context

roost runs on the user's own machines: WSL on Windows, Ubuntu VPSes, Raspberry Pis (ARM64 and 32-bit ARM). The first prototype was Bash + PowerShell. It was quick for discovery, but hard to test, fragile with JSON and hard to ship to strangers.

## Decision

- **Go**, built as one static binary per platform (`CGO_ENABLED=0`), released by GoReleaser.
- **cobra** for commands. **No viper.** Config is 4 keys in one TOML file (BurntSushi/toml).
- PowerShell stays, but only as an embedded asset for the one admin-elevated Windows action.

## Alternatives

| Option | Why not |
|---|---|
| Bash | no real tests, JSON parsing needs jq/python, quoting bugs, hard to grow (sync, vault, dashboard) |
| Python | users need the right interpreter + packages; awkward single-file shipping, worse on Pi/Windows |
| C# | great on Windows, but roost is mostly Linux |
| Rust | would work; slower to write and change for a tool that mostly configures other tools |

Go is the language of the tools roost builds on (Tailscale, Syncthing, restic, Docker), cross-compiles trivially, and has a stdlib HTTP server for the future dashboard.

## Consequences

- Contributors need Go 1.27+.
- No native Windows build: on Windows, roost runs inside WSL.
- Adding viper later is possible if config grows layers (env/flags/file); not needed now.
