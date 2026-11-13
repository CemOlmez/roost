# .github/ — CI, security, releases

## What lives here

- **`workflows/`**: GitHub Actions. See `workflows/README.md`.
- **`ISSUE_TEMPLATE/`**: the setup-problem form (asks for `roost doctor --json`).
- **`dependabot.yml`**: weekly grouped updates for Go modules and Actions.

## Naming

GitHub's fixed names.

## Read first if

CI fails, or you're changing how releases are built.

## Skip / don't read

Everything here for code changes.

## Discipline

- Workflows get least-privilege `permissions:`; only `release.yml` may write contents.

## Cross-references

- **Consumes:** `Makefile` targets mirror CI; `.goreleaser.yaml`; `.golangci.yml`.
- **Feeds:** GitHub Releases (what `install.sh` downloads).
