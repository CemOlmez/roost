# No Roost server; build on Tailscale

**Date:** 2026-11-13 · **Status:** accepted

## Context

A tool that reaches into someone's home PC from anywhere is a security product whether it means to be or not. A Roost-run server (accounts, relay, dashboard) would hold the keys to every user's machine.

## Decision

- roost has **no server, no account, no telemetry**. Everything runs on the user's machines.
- Networking is **Tailscale**: devices authenticate with the user's tailnet identity, and nothing listens on the public internet. Shell access is **Tailscale SSH** (no OpenSSH server to expose or harden).
- Phone access is **Claude Code Remote Control**, which connects outbound to Anthropic.
- Future off-site backup (vault) goes to a bucket **the user owns**, encrypted with a key only they hold.

## Alternatives

- A hosted relay/dashboard: easier onboarding, but central risk, cost and lock-in.
- Raw WireGuard / Headscale: possible later for people who won't use Tailscale; more setup now.

## Consequences

- Users need a (free) Tailscale account.
- "Sign in on a friend's laptop" means signing in to Tailscale + entering the vault passphrase, not a Roost login.
