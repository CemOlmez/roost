# docs/ — how roost works and why

Explanation, not tracking. Open work lives in `tasks/`; what happened lives in `NOTES.md`.

## What lives here

- **`DESIGN.md`**: the product vision: concepts (box, primary, replica, vault, waker), real-world scenarios, security model, hard problems, roadmap. Written before the code; describes v0.1 *and* the future.
- **`ARCHITECTURE.md`**: how the code is built today: packages, the step model, adding a step, testing, releases.
- **`workflow.md`**: how work moves through this repo (intake → build → gate → trackers).
- **`decisions/`**: one file per irreversible choice. See `decisions/README.md`.

## Naming

Standing docs have fixed names. Decisions are `decisions/YY-MM-DD_<topic>.md`.

## Read first if

ARCHITECTURE: you're changing code. DESIGN: you're planning a roadmap feature. workflow: you're unsure of the process.

## Skip / don't read

DESIGN.md for bug fixes and step changes; it's long and mostly about future versions.

## Discipline

- When code and ARCHITECTURE.md disagree, the code is right; fix the doc in the same change.
- DESIGN.md features that ship move to README.md (user view) and ARCHITECTURE.md (code view).

## Cross-references

- **Consumes:** `internal/` (ARCHITECTURE describes it).
- **Feeds:** `tasks/backlog.md` (roadmap items), `README.md`.
