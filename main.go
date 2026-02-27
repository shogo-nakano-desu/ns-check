package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shogonakano/nmchk/checker"
	"github.com/shogonakano/nmchk/output"
	"github.com/shogonakano/nmchk/runner"
)

var version = "0.1.0"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("nmchk", flag.ContinueOnError)

	only := fs.String("only", "", "comma-separated list of registries to check (e.g. npm,github)")
	skip := fs.String("skip", "", "comma-separated list of registries to skip (e.g. domain)")
	timeout := fs.Duration("timeout", 10*time.Second, "timeout for all checks")
	noColor := fs.Bool("no-color", false, "disable colored output")
	showVersion := fs.Bool("version", false, "print version and exit")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "nmchk - NaMe CHecK. Check namespace availability.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: nmchk [flags] <name>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nRegistries: domain (.com/.io/.net/.app/.ai/.sh/.tech), npm, github, github-repo, dockerhub, crates, homebrew\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  nmchk myproject\n")
		fmt.Fprintf(os.Stderr, "  nmchk --only npm,github myproject\n")
		fmt.Fprintf(os.Stderr, "  nmchk --skip domain myproject\n")
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Printf("nmchk v%s\n", version)
		return 0
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	name := fs.Arg(0)

	if *only != "" && *skip != "" {
		fmt.Fprintln(os.Stderr, "error: --only and --skip are mutually exclusive")
		return 2
	}

	allCheckers := buildCheckers()
	filtered := filterCheckers(allCheckers, *only, *skip)

	if len(filtered) == 0 {
		fmt.Fprintln(os.Stderr, "error: no registries selected after filtering")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	results := runner.Run(ctx, filtered, name)

	useColor := output.ShouldUseColor(*noColor)
	printer := output.NewPrinter(useColor)
	printer.Print(name, results)

	return exitCode(results)
}

func buildCheckers() []checker.Checker {
	client := &http.Client{}
	ghToken := os.Getenv("GITHUB_TOKEN")

	domainTLDs := []string{"com", "io", "net", "app", "ai", "sh", "tech"}
	checkers := make([]checker.Checker, 0, len(domainTLDs)+6)
	for _, tld := range domainTLDs {
		checkers = append(checkers, checker.NewDefaultDomainChecker(tld))
	}

	checkers = append(checkers,
		checker.NewNpmChecker(client, "https://registry.npmjs.org"),
		checker.NewCratesChecker(client, "https://crates.io"),
		checker.NewGitHubChecker(client, "https://api.github.com", ghToken),
		checker.NewGitHubRepoChecker(client, "https://api.github.com", ghToken),
		checker.NewDockerHubChecker(client, "https://hub.docker.com"),
		checker.NewHomebrewChecker(client, "https://formulae.brew.sh"),
	)
	return checkers
}

func filterCheckers(all []checker.Checker, only, skip string) []checker.Checker {
	if only != "" {
		set := toSet(only)
		var out []checker.Checker
		for _, c := range all {
			if set[c.Name()] {
				out = append(out, c)
			}
		}
		return out
	}
	if skip != "" {
		set := toSet(skip)
		var out []checker.Checker
		for _, c := range all {
			if !set[c.Name()] {
				out = append(out, c)
			}
		}
		return out
	}
	return all
}

func toSet(csv string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range strings.Split(csv, ",") {
		m[strings.TrimSpace(strings.ToLower(s))] = true
	}
	return m
}

func exitCode(results []checker.Result) int {
	hasError := false
	for _, r := range results {
		switch r.Status {
		case checker.Taken:
			return 1
		case checker.Unknown:
			hasError = true
		}
	}
	if hasError {
		return 2
	}
	return 0
}
