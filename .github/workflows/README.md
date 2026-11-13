# .github/workflows/ — GitHub Actions

| Workflow | Trigger | Does |
|---|---|---|
| `ci.yml` | push to main, PRs | tidy check, build, `go test -race`, golangci-lint, shellcheck, goreleaser snapshot |
| `security.yml` | push, PRs, weekly | govulncheck, CodeQL (security-extended) |
| `release.yml` | tag `v*` | tests, then GoReleaser → GitHub Release with `roost_linux_<arch>.tar.gz` + `checksums.txt` |

## Discipline

- Pin major versions of actions; Dependabot bumps them.
- Keep `make ci` and `ci.yml` doing the same things.
