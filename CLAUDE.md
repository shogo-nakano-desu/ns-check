# CLAUDE.md

namo is a Go CLI tool that checks namespace availability across package registries, domain names, and platforms in parallel.

## Quick Reference

- **Language:** Go 1.25+ (stdlib only, zero external dependencies)
- **Build:** `go build -o namo .`
- **Test (unit):** `go test ./...`
- **Test (e2e):** `go test -tags e2e ./...`
- **Lint:** `go vet ./...`

## Documents

- [README.md](README.md) — Usage, install instructions, registry list, exit codes, environment variables
- [CONTRIBUTING.md](CONTRIBUTING.md) — Dev setup, testing strategy, how to add a new checker, commit conventions, PR guidelines
- [PLAN.md](PLAN.md) — Original design spec: architecture, implementation steps, core types, registry check details

## Project Layout

```
main.go              CLI entry point, flag parsing, checker wiring
checker/
  checker.go         Checker interface, Result type, Status enum
  domain.go          DNS lookup for <name>.com
  npm.go             npm registry
  github.go          GitHub user/org
  github_repo.go     GitHub repository search
  dockerhub.go       Docker Hub namespace
  crates.go          crates.io (Rust)
  homebrew.go        Homebrew formulae & casks
  *_test.go          Unit tests (httptest servers, no network)
runner/
  runner.go          Concurrent execution engine (goroutines + channel)
  runner_test.go     Runner tests with mock checkers
output/
  output.go          Terminal output formatting, ANSI colors, TTY detection
  output_test.go     Output tests with bytes.Buffer
e2e_test.go          End-to-end tests (build tag: e2e for real network)
```

## Key Patterns

- **Checker interface:** `Name()`, `DisplayName()`, `Check(ctx, name) Result` — all checkers implement this
- **Dependency injection:** HTTP clients and base URLs are injectable for testability
- **Testing:** Unit tests use `httptest.NewServer` (no network); e2e tests behind `//go:build e2e`
- **Three-state results:** Available, Taken, or Unknown (never silently swallow errors)
- **Concurrency:** All checks run in parallel via `runner.Run()`

## Conventions

- **Commits:** [Conventional Commits](https://www.conventionalcommits.org/) — `feat`, `fix`, `docs`, `test`, `refactor`, `chore`
- **No external deps:** Everything uses Go stdlib
- **One PR, one thing:** New checker, bug fix, or refactor — not all three
