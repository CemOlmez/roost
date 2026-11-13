# .claude/ — shared agent configuration

## What lives here

- **`skills/`**: Go skills vendored from [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang) (MIT, see `skills/LICENSE-cc-skills-golang`): `golang-cli`, `golang-project-layout`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-continuous-integration`. Loaded automatically by Claude Code in this repo.

## Read first if

You're working on the CLI (`golang-cli`), tests (`golang-testing`) or CI (`golang-continuous-integration`).

## Skip / don't read

The skills' `references/` files unless a skill points you there.

## Discipline

- Local-only settings (`settings.local.json`) are gitignored. Never commit them.
- Update vendored skills deliberately: re-read the diff before replacing.
