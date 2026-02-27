package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "nmchk-e2e")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	binaryPath = filepath.Join(dir, "nmchk")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	os.Exit(m.Run())
}

func runNmchk(args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestE2E_NoArgs(t *testing.T) {
	_, stderr, code := runNmchk()

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected usage in stderr, got:\n%s", stderr)
	}
}

func TestE2E_Version(t *testing.T) {
	stdout, _, code := runNmchk("--version")

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout, "nmchk v") {
		t.Errorf("expected version output, got:\n%s", stdout)
	}
}

func TestE2E_OnlyAndSkipMutuallyExclusive(t *testing.T) {
	_, stderr, code := runNmchk("--only", "npm", "--skip", "domain", "test")

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got:\n%s", stderr)
	}
}

func TestE2E_UnknownFlag(t *testing.T) {
	_, stderr, code := runNmchk("--nonexistent", "test")

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if stderr == "" {
		t.Error("expected error output for unknown flag")
	}
}

func TestE2E_NoRegistriesAfterFilter(t *testing.T) {
	_, stderr, code := runNmchk("--only", "nonexistent", "test")

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, "no registries selected") {
		t.Errorf("expected 'no registries selected' error, got:\n%s", stderr)
	}
}

func TestE2E_OnlyFilter(t *testing.T) {
	stdout, _, _ := runNmchk("--only", "npm", "--no-color", "react")

	if !strings.Contains(stdout, "npm") {
		t.Errorf("expected 'npm' in output, got:\n%s", stdout)
	}
	// Should NOT contain any other registries
	for _, excluded := range []string{"GitHub", "Docker Hub", "Domain", "crates.io", "Homebrew"} {
		if strings.Contains(stdout, excluded) {
			t.Errorf("expected no %s in --only npm output, got:\n%s", excluded, stdout)
		}
	}
}

func TestE2E_SkipFilter(t *testing.T) {
	stdout, _, _ := runNmchk("--skip", "domain,dockerhub", "--no-color", "react")

	// Should contain the non-skipped registries
	for _, expected := range []string{"npm", "crates.io", "GitHub", "Homebrew"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("expected '%s' in output, got:\n%s", expected, stdout)
		}
	}
	// Should NOT contain skipped registries
	if strings.Contains(stdout, "Domain") {
		t.Errorf("expected no Domain in --skip domain output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "Docker Hub") {
		t.Errorf("expected no Docker Hub in --skip dockerhub output, got:\n%s", stdout)
	}
}

func TestE2E_NoColorFlag(t *testing.T) {
	stdout, _, _ := runNmchk("--only", "npm", "--no-color", "react")

	if strings.Contains(stdout, "\033[") {
		t.Error("expected no ANSI escape codes with --no-color")
	}
}

func TestE2E_TakenName_ExitCode1(t *testing.T) {
	// "react" is taken on npm
	_, _, code := runNmchk("--only", "npm", "react")

	if code != 1 {
		t.Errorf("expected exit code 1 for taken name, got %d", code)
	}
}

func TestE2E_TimeoutTriggersError(t *testing.T) {
	// 1ms timeout should cause all checks to fail
	stdout, _, code := runNmchk("--timeout", "1ms", "--no-color", "test")

	// Should be exit 2 (errors) or possibly exit 1 if domain resolved from cache
	if code == 0 {
		t.Errorf("expected non-zero exit code with 1ms timeout, got 0.\nOutput:\n%s", stdout)
	}
}

func TestE2E_OutputContainsAvailableCount(t *testing.T) {
	stdout, _, _ := runNmchk("--only", "npm", "--no-color", "react")

	if !strings.Contains(stdout, "of 1 available") {
		t.Errorf("expected 'of 1 available' in output, got:\n%s", stdout)
	}
}

func TestE2E_CratesChecker(t *testing.T) {
	// "serde" is a well-known Rust crate
	stdout, _, code := runNmchk("--only", "crates", "--no-color", "serde")

	if code != 1 {
		t.Errorf("expected exit code 1 for taken crate, got %d", code)
	}
	if !strings.Contains(stdout, "crates.io") {
		t.Errorf("expected 'crates.io' in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "taken") {
		t.Errorf("expected 'taken' in output, got:\n%s", stdout)
	}
}

func TestE2E_GitHubRepoChecker(t *testing.T) {
	// "react" is a well-known GitHub repo
	stdout, _, code := runNmchk("--only", "github-repo", "--no-color", "react")

	if code != 1 {
		t.Errorf("expected exit code 1 for taken repo, got %d", code)
	}
	if !strings.Contains(stdout, "GitHub Repo") {
		t.Errorf("expected 'GitHub Repo' in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "taken") {
		t.Errorf("expected 'taken' in output, got:\n%s", stdout)
	}
}

func TestE2E_HomebrewChecker(t *testing.T) {
	// "wget" is a well-known Homebrew formula
	stdout, _, code := runNmchk("--only", "homebrew", "--no-color", "wget")

	if code != 1 {
		t.Errorf("expected exit code 1 for taken formula, got %d", code)
	}
	if !strings.Contains(stdout, "Homebrew") {
		t.Errorf("expected 'Homebrew' in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "taken") {
		t.Errorf("expected 'taken' in output, got:\n%s", stdout)
	}
}

func TestE2E_AllRegistriesOutput(t *testing.T) {
	stdout, _, _ := runNmchk("--no-color", "react")

	// All 13 registries should appear in output (7 domain TLDs + 6 others)
	for _, reg := range []string{
		"Domain (.com)", "Domain (.io)", "Domain (.net)", "Domain (.app)",
		"Domain (.ai)", "Domain (.sh)", "Domain (.tech)",
		"npm", "crates.io", "GitHub", "GitHub Repo", "Docker Hub", "Homebrew",
	} {
		if !strings.Contains(stdout, reg) {
			t.Errorf("expected '%s' in output, got:\n%s", reg, stdout)
		}
	}
	if !strings.Contains(stdout, "of 13 available") {
		t.Errorf("expected 'of 13 available' in output, got:\n%s", stdout)
	}
}

func TestE2E_MultipleArgs(t *testing.T) {
	_, stderr, code := runNmchk("name1", "name2")

	if code != 2 {
		t.Errorf("expected exit code 2 for multiple args, got %d", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected usage in stderr, got:\n%s", stderr)
	}
}
