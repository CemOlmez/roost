# Contributing to Roost

Thanks for helping! The most valuable contribution right now is **a report from a real setup**: what worked, what didn't, on what hardware.

## Reporting a problem

Open an issue with:
- your machine (Windows + WSL? VPS? Pi model?) and Linux distro
- the output of `roost doctor --json`
- what you expected vs. what happened

## Development

Requirements: Go 1.27+, [golangci-lint](https://golangci-lint.run) v2, shellcheck.

```bash
git clone https://github.com/CemOlmez/roost && cd roost
make ci          # tidy + lint + test + shellcheck + build — must pass before a PR
./bin/roost doctor   # read-only, safe on your machine
```

- How the code is organized and how to add a setup step: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- How work flows (tests first, docs updated in the same change): [docs/workflow.md](docs/workflow.md)

## Rules that keep roost trustworthy

- All OS access goes through `internal/host`, and every step has a test using `hosttest.Fake`.
- Every change roost makes to a machine is listed in the README with how to undo it.
- roost never handles passwords and never phones home.
- Windows checks must not parse localized text (read registry values and hex values).

## Pull requests

- One logical change per PR; conventional commit titles (`feat:`, `fix:`, `docs:`, `ci:`).
- Add a line to `CHANGELOG.md` under **Unreleased** for user-visible changes.
