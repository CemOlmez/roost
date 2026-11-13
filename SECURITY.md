# Security Policy

roost configures remote access to people's machines, so security reports are taken seriously.

## Reporting a vulnerability

Please **don't open a public issue**. Use GitHub's private reporting instead: **Security → Report a vulnerability** on this repository. You'll get a reply within a week.

## Scope

In scope: anything that lets roost expose a machine beyond the user's tailnet, run unintended commands, weaken file permissions, leak secrets, or install something other than what was asked (including `install.sh` and release integrity).

## How the box can be reached

A roost box has **two** ways in. Protect both:

1. **Your tailnet** (Tailscale SSH). Who can log in, and as which user, is set by your tailnet's SSH policy in the Tailscale admin console. On a personal tailnet, we recommend:
   ```json
   "ssh": [{ "action": "check", "src": ["autogroup:member"], "dst": ["autogroup:self"], "users": ["autogroup:nonroot"] }]
   ```
   Don't run a box on a shared or company tailnet unless you have checked its SSH rules.
2. **Your Claude account** (Remote Control). Anyone who controls your claude.ai account can start sessions on the box, which means they can run commands as you. Turn on 2-step verification or SSO. To shut this path off: `roost rc stop`, and `systemctl --user disable roost-rc` to keep it off.

## Design guarantees

- No roost server, account or telemetry.
- Nothing listens on the public internet. Shell access goes through Tailscale SSH, and Remote Control connects outbound.
- roost never sees or stores passwords; sudo and UAC prompts go to the user.
- The Windows admin step runs a script embedded in the roost binary. It is passed to PowerShell in memory, never from a file another process could swap.
- Releases are built by GitHub Actions from tagged commits, with pinned actions. `install.sh` verifies SHA-256 checksums: this catches corrupted downloads, but the checksums come from the same release, so it is an integrity check, not proof of authorship. For that, each release carries a signed build-provenance attestation: `gh attestation verify roost_linux_amd64.tar.gz --repo CemOlmez/roost`.
