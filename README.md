# namo

**Is that name taken?** Find out everywhere at once.

namo checks namespace availability across package registries, domain names, and platforms — all in a single command. Stop manually visiting seven different websites before naming your next project.

```
$ namo aurora

  ✗  domain          taken (76.76.21.21)
  ✓  npm             available
  ✗  github          taken (Organization)
  ✗  github-repo     taken (nicehash/aurora)
  ✗  dockerhub       taken
  ✓  crates          available
  ✗  homebrew        taken (formula)

  2 of 7 available
```

## Features

- **7 registries** — domain (.com), npm, GitHub user/org, GitHub repo, Docker Hub, crates.io, Homebrew
- **Parallel checks** — all registries queried concurrently, results in seconds
- **Zero dependencies** — single static binary, built with Go stdlib only
- **Smart exit codes** — scriptable: `0` all available, `1` some taken, `2` error
- **Filterable** — `--only` and `--skip` flags to check exactly what you need
- **Color-aware** — respects `NO_COLOR`, auto-detects TTY, `--no-color` flag

## Install

### From source

```sh
go install github.com/shogonakano/namo@latest
```

### Build locally

```sh
git clone https://github.com/shogonakano/namo.git
cd namo
go build -o namo .
```

## Usage

```sh
# Check all registries
namo myproject

# Check only npm and crates.io
namo --only npm,crates myproject

# Skip domain lookup
namo --skip domain myproject

# Custom timeout (default: 10s)
namo --timeout 5s myproject

# Disable colors
namo --no-color myproject

# Use in scripts
namo coolname && echo "It's all yours!"
```

### Registries

| Name          | What it checks                          |
|---------------|-----------------------------------------|
| `domain`      | DNS lookup for `<name>.com`             |
| `npm`         | npm registry                            |
| `github`      | GitHub username / organization          |
| `github-repo` | GitHub repository (exact name match)    |
| `dockerhub`   | Docker Hub namespace                    |
| `crates`      | Rust crates.io                          |
| `homebrew`    | Homebrew formulae and casks             |

### Exit codes

| Code | Meaning                              |
|------|--------------------------------------|
| `0`  | All checked registries are available |
| `1`  | At least one registry is taken       |
| `2`  | Error (bad input, timeout, etc.)     |

### Environment variables

| Variable       | Description                                              |
|----------------|----------------------------------------------------------|
| `GITHUB_TOKEN` | GitHub personal access token for higher API rate limits  |
| `NO_COLOR`     | Set to any value to disable colored output               |

## Architecture

```
main.go            CLI entry point, flag parsing
checker/           Registry checker implementations (one file per registry)
  checker.go       Checker interface & shared types
runner/            Concurrent execution engine
output/            Terminal output formatting
```

Each checker implements a simple interface:

```go
type Checker interface {
    Name() string                                    // registry key for --only/--skip
    DisplayName() string                             // human-readable label
    Check(ctx context.Context, name string) Result   // the actual check
}
```

Adding a new registry is just implementing this interface and registering it in `main.go`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and how to add new checkers.

## License

MIT
