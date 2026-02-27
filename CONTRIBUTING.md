# Contributing to namo

Thanks for wanting to help! namo is a small, focused tool and contributions of all sizes are welcome — from typo fixes to new registry checkers.

## Development setup

**Requirements:** Go 1.25+

```sh
git clone https://github.com/shogonakano/namo.git
cd namo
go build -o namo .
```

## Running tests

```sh
# Unit tests (fast, no network)
go test ./...

# Include end-to-end tests (hits real APIs)
go test -tags e2e ./...
```

All unit tests use mock HTTP servers — no network access required. E2E tests make real API calls and are gated behind the `e2e` build tag.

## Project structure

```
main.go              CLI entry point & flag parsing
e2e_test.go          End-to-end tests (build tag: e2e)
checker/
  checker.go         Checker interface, Result, Status types
  domain.go          Domain (.com) availability via DNS
  npm.go             npm registry
  github.go          GitHub user/org
  github_repo.go     GitHub repository search
  dockerhub.go       Docker Hub namespace
  crates.go          Rust crates.io
  homebrew.go        Homebrew formula & cask
  *_test.go          Unit tests for each checker
runner/
  runner.go          Concurrent checker execution
  runner_test.go     Runner tests
output/
  output.go          Terminal output formatting & colors
  output_test.go     Output tests
```

## Adding a new registry checker

This is the most common type of contribution. Here's the process:

### 1. Create the checker

Add `checker/<registry>.go`:

```go
package checker

import (
    "context"
    "net/http"
)

type MyRegistryChecker struct {
    client  *http.Client
    baseURL string // injectable for testing
}

func NewMyRegistryChecker(client *http.Client) *MyRegistryChecker {
    return &MyRegistryChecker{client: client, baseURL: "https://api.myregistry.com"}
}

func (c *MyRegistryChecker) Name() string        { return "myregistry" }
func (c *MyRegistryChecker) DisplayName() string  { return "My Registry" }

func (c *MyRegistryChecker) Check(ctx context.Context, name string) Result {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/"+name, nil)
    if err != nil {
        return Result{Registry: c.DisplayName(), Name: name, Status: StatusUnknown, Err: err}
    }
    req.Header.Set("User-Agent", "namo/1.0")

    resp, err := c.client.Do(req)
    if err != nil {
        return Result{Registry: c.DisplayName(), Name: name, Status: StatusUnknown, Err: err}
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case http.StatusOK:
        return Result{Registry: c.DisplayName(), Name: name, Status: StatusTaken}
    case http.StatusNotFound:
        return Result{Registry: c.DisplayName(), Name: name, Status: StatusAvailable}
    default:
        return Result{Registry: c.DisplayName(), Name: name, Status: StatusUnknown,
            Err: fmt.Errorf("unexpected status: %d", resp.StatusCode)}
    }
}
```

### 2. Write tests

Add `checker/<registry>_test.go` using `httptest.NewServer`:

```go
package checker

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestMyRegistryChecker_Taken(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer srv.Close()

    c := &MyRegistryChecker{client: srv.Client(), baseURL: srv.URL}
    result := c.Check(context.Background(), "taken-name")

    if result.Status != StatusTaken {
        t.Errorf("got %v, want StatusTaken", result.Status)
    }
}
```

Cover at minimum: taken, available, server error, timeout, and correct URL/headers.

### 3. Register it

In `main.go`, add to the checkers slice:

```go
checker.NewMyRegistryChecker(httpClient),
```

### 4. Verify

```sh
go test ./...
go build -o namo .
./namo testname
```

## Design principles

- **No external dependencies.** Everything uses Go stdlib. Don't add third-party packages.
- **Inject for testing.** HTTP clients and base URLs are always injectable so tests use `httptest.NewServer` — no network calls in unit tests.
- **Three-state results.** Every check returns Available, Taken, or Unknown. Never silently swallow errors.
- **Respect rate limits.** Handle 429/403 gracefully. Support auth tokens where the API offers them.
- **Keep it fast.** All checks run concurrently. Don't add blocking sequential operations.

## Commit messages

This project follows [Conventional Commits](https://www.conventionalcommits.org/). Every commit message must be structured as:

```
<type>[optional scope]: <description>

[optional body]
```

**Types:**

| Type       | When to use                                    |
|------------|------------------------------------------------|
| `feat`     | A new feature (new checker, new flag, etc.)    |
| `fix`      | A bug fix                                      |
| `docs`     | Documentation only changes                     |
| `test`     | Adding or updating tests                       |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `chore`    | Build process, CI, tooling changes             |

**Examples:**

```
feat(checker): add PyPI registry checker
fix(runner): prevent goroutine leak on context cancellation
docs: update README with install instructions
test(homebrew): add case for formula-only match
refactor(output): extract color logic into helper
```

Keep the subject line under 72 characters. Use the body for "why", not "what".

## Pull request guidelines

1. **One thing per PR.** A new checker, a bug fix, or a refactor — not all three.
2. **Tests required.** All new code needs tests. All tests must pass.
3. **Run `go vet ./...`** before submitting.
4. **Conventional commits.** Follow the commit message format described above.
5. **Keep the interface small.** namo is intentionally simple. If a feature needs a flag, think twice.

## Reporting bugs

Open an issue with:

- What you ran (command + flags)
- What you expected
- What actually happened
- Go version (`go version`)

## Ideas for contributions

- New registry checkers (PyPI, RubyGems, Maven, etc.)
- Better error messages for common failure modes
- Shell completions (bash, zsh, fish)
- Homebrew formula for installing namo itself
