# docs/decisions/ — irreversible choices

One file per decision that would be expensive to reverse: language, distribution, network layer, data model. Small or reversible choices go in `NOTES.md` only.

## What lives here

- **`26-11-13_go-single-binary.md`**: why Go (and not Bash, Python, C# or Rust), and why cobra without viper.
- **`26-11-13_no-roost-server.md`**: why roost has no server, account or telemetry, and builds on Tailscale.

## Naming

`YY-MM-DD_<topic>.md`. Sections: Context · Decision · Alternatives · Consequences.

## Read first if

You're about to change something one of these files covers.

## Skip / don't read

Everything else, for routine work.

## Discipline

- Never edit a decision to change it. Write a new one that supersedes it and link both ways.

## Cross-references

- **Consumes:** `NOTES.md` (decisions start there).
- **Feeds:** `tasks/handoff.md` § Locked decisions.
