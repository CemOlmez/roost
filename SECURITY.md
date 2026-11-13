# Security Policy

roost configures remote access to people's machines, so security reports are taken seriously.

## Reporting a vulnerability

Please **don't open a public issue**. Use GitHub's private reporting instead: **Security → Report a vulnerability** on this repository. You'll get a reply within a week.

## Scope

In scope: anything that lets roost expose a machine beyond the user's tailnet, run unintended commands, weaken file permissions, leak secrets, or install something other than what was asked (including `install.sh` and release integrity).

## Design guarantees

- No roost server, account or telemetry.
- Shell access only through Tailscale SSH (tailnet identity); nothing listens publicly.
- roost never sees or stores passwords; sudo and UAC prompts go to the user.
- Releases are built by GitHub Actions from tagged commits; `install.sh` verifies SHA-256 checksums before installing.
