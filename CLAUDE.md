# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Roost** is an open-source, single-binary Go CLI that turns a machine the user owns (a Windows PC with WSL, a VPS, a Raspberry Pi) into an always-on Linux dev box, reachable from a phone through Claude Code Remote Control and from a laptop through SSH. It recovers by itself after reboots and power cuts. v1 means a stranger can run `install.sh`, follow `roost init`, and close their laptop to keep coding from their phone.

Framings that must not be gotten wrong:
- **Roost is local software, not a SaaS.** There is no Roost server, account or telemetry. Anything that adds one is a design change, not a feature.
- **Roost glues proven tools together; it doesn't replace them.** It uses Tailscale, tmux, systemd, Claude Code and (later) Syncthing and restic. Don't reimplement networking, sync or backup.
- **Two roles only: `box` (runs the work 24/7) and `client` (a laptop that connects).** "Primary/replica" in DESIGN.md is the *future* sync model (v0.3), not current code.
- **Roost never handles the user's passwords or secrets.** sudo prompts go to the terminal; Windows auto-login is a manual step.
- **The box runs Linux.** On Windows, roost runs *inside* WSL and reaches Windows through `powershell.exe`. There is no native Windows build.

## Repository Structure

- **`cmd/roost/`**: `main`. Wires the real host into the CLI. See `cmd/README.md`.
- **`internal/`**: all logic. See `internal/README.md` for the package inventory.
  - **`internal/steps/`**: the setup checklist (what init/doctor do). See `internal/steps/README.md` for the step inventory.
- **`docs/`**: DESIGN (vision), ARCHITECTURE (how the code works), workflow, decisions. See `docs/README.md`.
- **`tasks/`**: backlog and session handoff. See `tasks/README.md`.
- **`.github/`**: CI, security scans, release, Dependabot. See `.github/README.md`.
- **`.claude/skills/`**: vendored Go skills (samber/cc-skills-golang, MIT): cli, project-layout, testing, error-handling, security, continuous-integration.
- **`install.sh`**: the curl-pipe installer; downloads a release and verifies its checksum.
- **`STATE.md`**: current snapshot. Rewritten in place, never appended.
- **`NOTES.md`**: dated journal, newest first. Also holds real-hardware test logs.
- **`README.md`**: user-facing. **`CONTRIBUTING.md`**, **`SECURITY.md`**, **`CHANGELOG.md`**: standard OSS files.

## Master Routing Table

| Task | Read First | Then Read | Skip |
|---|---|---|---|
| Start any session | `STATE.md` (full), `NOTES.md` (top 2 entries), `tasks/backlog.md` (`## MVP`) | `docs/workflow.md` if the process is unclear | rest of NOTES, `docs/DESIGN.md`, `.claude/skills/` |
| Add or change a setup step | `internal/steps/README.md` (inventory), `docs/ARCHITECTURE.md` § Adding a step | `internal/step/step.go` | `internal/cli/`, `docs/DESIGN.md` |
| Add or change a command | `internal/cli/README.md`, `.claude/skills/golang-cli/SKILL.md` | `internal/step/run.go` | `internal/steps/assets/` |
| Windows-side behaviour | `internal/steps/README.md`, `internal/steps/windows.go`, `internal/steps/assets/roost-windows.ps1` | `internal/host/platform.go` | `internal/tailscale/`, `internal/ini/` internals |
| Fix a bug from a real setup | `NOTES.md` (matching test log), `internal/steps/README.md` | the step's test in `steps_test.go` | `docs/DESIGN.md` |
| Installer / release | `install.sh`, `.goreleaser.yaml`, `.github/README.md` | `docs/ARCHITECTURE.md` § Releases | `internal/` |
| Plan a roadmap feature (wake, sync, vault) | `docs/DESIGN.md`, `tasks/backlog.md` | `docs/decisions/` | `internal/cli/`, assets |
| User-facing docs | `README.md` | `docs/ARCHITECTURE.md` | `internal/`, `.claude/` |

## File Naming Conventions

| Type | Pattern | Example |
|---|---|---|
| Go source | `lowercase.go`, one area per file | `internal/steps/network.go` |
| Go test | `<file>_test.go` next to the code | `internal/ini/ini_test.go` |
| Test fixture | testdata/<what>.json beside the test | `internal/tailscale/testdata/status.json` |
| Step ID | kebab-case, stable forever (users type it) | `tailscale-connected`, `bios` |
| Embedded asset | `internal/steps/assets/roost-<what>.<ext>` | `roost-rc.service.tmpl` |
| Decision record | `docs/decisions/YY-MM-DD_<topic>.md` | `26-11-13_go-single-binary.md` |

**Exempt:** `README.md`, `CLAUDE.md`, `STATE.md`, `NOTES.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE`, `docs/DESIGN.md`, `docs/ARCHITECTURE.md`, `docs/workflow.md`.

## Conventions

- **Read the README of the folder you are about to work in, before working in it.** The routing table says which folder; the folder's README has the local rules.
- **Reuse before create.** Before creating any component, module, hook, endpoint, table, or utility, read the inventory table in that folder's `README.md` and grep the folder for near-matches. If something covers the need, use it. If something almost covers it, extend or compose it — a new prop or variant beats a near-duplicate file. Build new only when nothing fits, and say in one line why nothing fit.

  **The inventory is updated in the same change that adds the unit.** A new component without its inventory row is incomplete work, not work plus a TODO. If a unit is deleted or renamed, its row goes with it.
- **Backlog capture is immediate and agent-initiated.** The moment something is deferred, discovered, or ruled out of scope — mid-implementation, mid-review, mid-conversation — write it into `tasks/backlog.md` *at that moment*, before continuing. Do not wait to be asked. Do not batch it to end of session. Triggers that mean "write a backlog item now": you say or think "we can do that later" / "out of scope" / "TODO" / "for now we'll…"; you work around a limitation instead of fixing it; you notice a bug or gap you are not fixing now; the user says "not now"; a decision is deferred to someone not in the session. At the end of every implementation, re-read what you built and extract remaining deferrals before reporting done.
- **All OS access goes through `host.Host`.** No exec.Command, os.ReadFile or os.Getenv outside `internal/host/`, because that is what keeps every step testable.
- **Every step change ships with a test** using `hosttest.Fake`. An unscripted command fails the test on purpose.
- **Every machine change is listed in README "What Roost changes on your machine"** with how to undo it. Users must be able to audit roost.
- **Windows checks parse registry or hex values, never localized words.** The author's Windows isn't in English; neither are many users'.
- **Decisions go in `NOTES.md` in real time**, with the reasoning. Irreversible ones also get a file in `docs/decisions/`.
- **Verification gate:** `make ci` must pass before claiming done. Quote the output.
- **Never commit secrets.** This repo is public.
- **No telemetry, no phoning home**, except downloading installers the user asked for (Tailscale, Claude Code) from their official URLs.

## Keeping This Current

The structure decays silently unless it is maintained in the same change that grows it.
These four fire automatically — do not wait to be asked:

- **New folder → its `README.md` in the same breath**, plus a line in the sitemap above
  and a routing row if it is somewhere sessions will work. A folder that appears without
  these is invisible to every future session.
- **New kind of task → a routing row.** If a task was not in the routing table, add it
  once the right Read/Skip answer is known. The table is only as good as its last update.
- **A convention discovered the hard way → the folder README's Discipline section**,
  before moving on. A trap hit twice is a trap that was never written down.
- **Renamed or deleted a unit → its inventory row goes with it.** A stale inventory is
  worse than none, because it is trusted.

Run `python3 ~/.claude/skills/project-bootstrap/scripts/check_structure.py .` to catch drift mechanically.

## Working Guidelines

Behavioural rules that cut the mistakes coding agents actually make. They bias toward
**caution over speed** — on a trivial task, use judgement. Where they collide with a
project rule above, the project rule wins.

### 1. Think before coding

Don't assume. Don't hide confusion. Surface tradeoffs.

- **State your assumptions explicitly before implementing.** If uncertain, ask.
- **If more than one reading of the request exists, present them** — never pick one
  silently and build on it.
- **If a simpler approach exists, say so.** Push back when it is warranted.
- **If something is unclear, stop and name what is confusing.** A question before the
  work costs one message; a wrong assumption costs the whole change.

### 2. Simplicity first

The minimum code that solves the problem. Nothing speculative.

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or configurability that was not requested.
- No error handling for scenarios that cannot occur.
- **If you wrote 200 lines and 50 would do, rewrite it before showing it.**

The test: *would a senior engineer call this overcomplicated?* If yes, simplify.

### 3. Surgical changes

Touch only what the task requires. Clean up only your own mess.

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor what isn't broken.
- Match the existing style, even where you would write it differently.
- Spotted unrelated dead code → **mention it, don't delete it.**
- Remove the imports, variables and functions that *your* change orphaned — and only
  those. Pre-existing dead code stays unless removing it was the task.

The test: **every changed line traces directly to the request.** A diff that is hard to
review is usually a diff that did more than it was asked to.

### 4. Goal-driven execution

Define success so it can be checked, then loop until it verifies.

Turn the task into a verifiable goal *before* starting:

| Asked for | Goal to work against |
|---|---|
| "Add validation" | write tests for the invalid inputs, then make them pass |
| "Fix the bug" | write a test that reproduces it, then make it pass |
| "Refactor X" | the tests pass before and after, unchanged |

For anything multi-step, state the plan first — one verification per step:

1. {step} → verify: {check}
2. {step} → verify: {check}

**Run the gate and quote the output. "Should pass" is not a result.** Strong success
criteria let you loop on your own; weak ones ("make it work") force the user to referee
every step.

---

**These are working if:** diffs carry fewer unrelated changes, less gets rewritten for
being overcomplicated, and the clarifying questions arrive *before* the implementation
rather than after the mistake.

## Commands

```bash
make build        # bin/roost for this machine
make test         # go test -race -shuffle=on ./...   (single test: go test ./internal/steps -run TestPackages)
make lint         # golangci-lint (also checks gofmt/goimports)
make ci           # the gate: tidy + lint + test + shellcheck + build
make snapshot     # local release build of all platforms (goreleaser)
./bin/roost doctor   # read-only; safe to run on the dev machine
```

Releasing: update `CHANGELOG.md`, then `git tag vX.Y.Z && git push origin vX.Y.Z`. CI builds and publishes the release.
