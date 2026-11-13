# Workflow

How work enters, moves through and leaves this repo.

## 1. Tiers

The core loop: **install → `roost init` → close the laptop → keep working from the phone → the box survives a reboot.**

- **MVP**: the loop breaks without it.
- **Core**: roost's identity (always reachable, never lost: wake, sync, vault).
- **Later**: an improvement, not a change of nature.

Tier inflation is the failure mode. If you can't name the loop step that breaks, it isn't MVP.

## 2. Idea intake

A new idea goes into `tasks/backlog.md` immediately, with its **ripple**: which steps, config keys, docs and README sections would have to change. Big ideas get a section in `docs/DESIGN.md` first.

## 3. The build loop

1. Read the folder README, and its inventory, for where you're working.
2. Reuse before create: extend a step, a host helper or a parser before writing a new one.
3. Write the test first (`hosttest.Fake` scripts exactly what the machine should see).
4. Build until it passes.
5. **Gate:** `make ci`. Quote the output.
6. Update the inventory row, README "What Roost changes", and `CHANGELOG.md` (Unreleased).
7. `NOTES.md` entry → extract deferrals into the backlog → delete the backlog item → update `STATE.md`.
8. Commit (conventional message: `feat:`, `fix:`, `docs:`, `ci:`, `chore:`).

## 4. The bug loop

1. Reproduce it: ideally a failing test in the step's `_test.go` using the real command output the user reported.
2. Decide whether the **doc** or the **code** is wrong.
3. A trap hit twice becomes a Discipline bullet in the folder README.
4. Anything that slipped past a green `make ci` gets a test.

## 5. Verification per kind of change

| Change | Verify with |
|---|---|
| Step logic | unit test with `hosttest.Fake`; then `./bin/roost doctor` on a real machine |
| Windows step / PowerShell | unit test of the parser + a real run on Windows (log in `NOTES.md`) |
| CLI command / flags | `internal/cli/cli_test.go` + run the binary |
| Installer | `make snapshot`, serve `dist/` locally, `ROOST_BASE_URL=… ROOST_BIN_DIR=… ROOST_NO_INIT=1 sh install.sh` |
| Release config | CI job "Release config builds" (goreleaser snapshot) |
| Docs only | links resolve; commands shown actually exist (`roost --help`) |

Never report "should pass".

## 6. Documentation contract: one job per file

| File | Job | Write mode |
|---|---|---|
| `README.md` | what users need: install, setup, commands, what changes | edited |
| `docs/ARCHITECTURE.md` | how the code works now | edited with the code |
| `docs/DESIGN.md` | vision and future | edited when plans change |
| `CHANGELOG.md` | user-visible changes per release | appended under Unreleased |
| `STATE.md` | what's true now | rewritten in place |
| `NOTES.md` | what happened, decisions, test logs | appended, newest first |
| `tasks/backlog.md` | what's not done | items added and deleted |
| folder `README.md` | local rules + inventory | edited with the folder |
