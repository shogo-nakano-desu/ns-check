# namo - NAMing OK? CLI Tool

## Context

When building a new product, checking namespace availability across multiple registries (domain, npm, GitHub, Docker Hub) is tedious and repetitive. `namo` is a Go CLI tool that checks all of them in parallel with a single command.

## Usage

```
namo myproject                       # check all registries
namo --only npm,github myproject     # check specific registries
namo --skip domain myproject         # skip specific registries
namo --timeout 5s myproject          # custom timeout
namo --no-color myproject            # disable colors
```

## Project Structure

```
namo/
├── go.mod
├── main.go                  # Entry point: flag parsing, orchestration, exit codes
├── checker/
│   ├── checker.go           # Checker interface + Result/Status types
│   ├── domain.go            # Domain (.com) via DNS lookup
│   ├── npm.go               # npm registry (HEAD https://registry.npmjs.org/<name>)
│   ├── github.go            # GitHub (GET https://api.github.com/users/<name>)
│   ├── dockerhub.go         # Docker Hub (GET https://hub.docker.com/v2/users/<name>)
│   ├── domain_test.go
│   ├── npm_test.go
│   ├── github_test.go
│   └── dockerhub_test.go
├── runner/
│   ├── runner.go            # Concurrent execution (goroutines + channel)
│   └── runner_test.go
└── output/
    ├── output.go            # Terminal output with ANSI colors
    └── output_test.go
```

## Core Design

### Checker Interface (`checker/checker.go`)

```go
type Status int
const (
    Available Status = iota
    Taken
    Unknown
)

type Result struct {
    Registry string
    Name     string
    Status   Status
    Err      error   // only when Status == Unknown
    Detail   string  // optional info (e.g. resolved IP)
}

type Checker interface {
    Name() string                                    // slug for --only/--skip (e.g. "npm")
    DisplayName() string                             // label for output (e.g. "npm registry")
    Check(ctx context.Context, name string) Result
}
```

### Registry Check Details

| Registry | Method | Endpoint | Available | Taken |
|----------|--------|----------|-----------|-------|
| Domain (.com) | `net.Resolver.LookupHost` | DNS for `<name>.com` | DNS NXDOMAIN | Returns addresses |
| npm | `HEAD` | `https://registry.npmjs.org/<name>` | 404 | 200 |
| GitHub | `GET` | `https://api.github.com/users/<name>` | 404 | 200 |
| Docker Hub | `GET` | `https://hub.docker.com/v2/users/<name>` | 404 | 200 |

- GitHub: reads `GITHUB_TOKEN` env var for authenticated requests (5000 req/hr vs 60 unauthenticated)
- All HTTP checkers: `User-Agent: namo/1.0` header
- Each checker struct accepts a base URL override for testability with `httptest.NewServer`

### Concurrency (`runner/runner.go`)

- Goroutines + buffered channel (capacity = number of checkers)
- `indexedResult` wrapper preserves original checker ordering
- `context.Context` flows to all checkers for timeout/cancellation
- Default timeout: 10 seconds

### Output (`output/output.go`)

```
Checking availability for "myproject"...

  Domain (.com)    ✓ available
  npm              ✗ taken
  GitHub           ✓ available
  Docker Hub       ✗ taken

2 of 4 available
```

- Green for available, red for taken, yellow for errors
- Auto-disable color when stdout is not a TTY (`os.ModeCharDevice` check)
- Respect `NO_COLOR` env var and `--no-color` flag

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All available |
| 1 | At least one taken |
| 2 | Error (or usage error) |

### CLI Flags (standard library `flag` package)

- `--only <registries>` — comma-separated, check only these
- `--skip <registries>` — comma-separated, skip these
- `--timeout <duration>` — default 10s
- `--no-color` — disable ANSI colors
- `--version` — print version
- `--only` and `--skip` are mutually exclusive

## Implementation Steps (TDD — tests first)

Each step writes the test(s) first, then implements the production code to make them pass.

### Step 1: Project init
- `go mod init`, create directory structure (`checker/`, `runner/`, `output/`)

### Step 2: Checker interface + types (`checker/checker.go`)
- Define `Status`, `Result`, `Checker` interface
- No tests needed (pure type definitions)

### Step 3: Runner — test first (`runner/`)
- **Write `runner_test.go` first**: mock `Checker` implementations with controlled delays and results
  - Test: all checkers complete, results returned in original order
  - Test: context cancellation propagates, returns Unknown
  - Test: concurrent execution (checkers with different delays all finish ~simultaneously)
- **Then implement `runner.go`**: goroutines + buffered channel + `indexedResult`

### Step 4: npm checker — test first (`checker/npm_test.go` → `checker/npm.go`)
- **Write tests first** using `httptest.NewServer`:
  - 200 → Taken
  - 404 → Available
  - 500 → Unknown with error
  - Context timeout → Unknown
- **Then implement** `NpmChecker` with `NewNpmChecker(client, baseURL)` constructor

### Step 5: GitHub checker — test first (`checker/github_test.go` → `checker/github.go`)
- **Write tests first** using `httptest.NewServer`:
  - 200 → Taken (parse `type` field for Detail)
  - 404 → Available
  - 403 with rate-limit headers → Unknown with rate-limit message
  - Token passed as Authorization header when provided
- **Then implement** `GitHubChecker`

### Step 6: Domain checker — test first (`checker/domain_test.go` → `checker/domain.go`)
- **Write tests first** with injectable `hostLookup` interface (fake resolver):
  - Returns addresses → Taken (Detail = resolved IPs)
  - Returns `*net.DNSError{IsNotFound: true}` → Available
  - Returns other error → Unknown
- **Then implement** `DomainChecker` with `net.Resolver` behind the interface

### Step 7: Docker Hub checker — test first (`checker/dockerhub_test.go` → `checker/dockerhub.go`)
- **Write tests first** using `httptest.NewServer`:
  - 200 → Taken
  - 404 → Available
  - 429 → Unknown (rate limited)
- **Then implement** `DockerHubChecker`

### Step 8: Output formatter — test first (`output/output_test.go` → `output/output.go`)
- **Write tests first** with `bytes.Buffer` as writer:
  - Test: available result shows "✓ available"
  - Test: taken result shows "✗ taken"
  - Test: error result shows "⚠" and error message
  - Test: summary line "N of M available"
  - Test: no-color mode strips ANSI codes
- **Then implement** `Printer` with ANSI color support, TTY detection

### Step 9: main.go — wire everything together
- Flag parsing with standard `flag` package
- `buildCheckers()`, `filterCheckers()`, `exitCode()`
- `--only`/`--skip` mutual exclusivity check
- Color auto-detection: `NO_COLOR` env + `os.ModeCharDevice` + `--no-color` flag

### Step 10: E2E tests (`e2e_test.go` in project root)
- **Build the binary** via `go build` in `TestMain`
- **Test CLI argument parsing**:
  - No args → exit 2, usage printed to stderr
  - `--version` → prints version, exit 0
  - `--only` and `--skip` together → exit 2, error message
  - Unknown flag → exit 2
- **Test with mock HTTP server** (set checkers' base URLs via env vars or test-only flags):
  - All registries available → exit 0
  - Some taken → exit 1
  - Error case → exit 2
- **Test real network calls** (behind `//go:build e2e` build tag, opt-in):
  - `namo react` → exit 1 (name widely taken)
  - `namo --only npm xyzzy-namo-unlikely-name-12345` → exit 0 (likely available)
  - `namo --timeout 1ms some-name` → exit 2 (timeout triggers errors)

## Testing Summary

| Package | Test File | Strategy |
|---------|-----------|----------|
| `checker/` | `npm_test.go` | `httptest.NewServer`, controlled HTTP responses |
| `checker/` | `github_test.go` | `httptest.NewServer`, rate-limit header testing |
| `checker/` | `domain_test.go` | Injectable `hostLookup` interface with fake resolver |
| `checker/` | `dockerhub_test.go` | `httptest.NewServer` |
| `runner/` | `runner_test.go` | Mock checkers with controlled delays |
| `output/` | `output_test.go` | `bytes.Buffer` writer injection |
| root | `e2e_test.go` | Build binary, `exec.Command`, assert stdout/stderr/exit code |
| root | `e2e_test.go` (`//go:build e2e`) | Real network calls against known names |

Run all unit tests: `go test ./...`
Run including e2e: `go test -tags e2e ./...`

## Verification

```bash
go test ./...                          # all unit tests pass
go build -o namo .                     # builds successfully
./namo react                           # exit 1, most registries taken
./namo xyzzy-namo-unlikely-12345       # exit 0, most registries available
./namo --only npm,github myproject     # filtered registries
./namo --skip domain myproject         # skip domain check
./namo --version                       # prints version
./namo                                 # exit 2, usage message
echo $?                                # verify exit codes
```
