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
	dir, err := os.MkdirTemp("", "namo-e2e")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	binaryPath = filepath.Join(dir, "namo")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	os.Exit(m.Run())
}

func runNamo(args ...string) (stdout, stderr string, exitCode int) {
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
	_, stderr, code := runNamo()

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected usage in stderr, got:\n%s", stderr)
	}
}

func TestE2E_Version(t *testing.T) {
	stdout, _, code := runNamo("--version")

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout, "namo v") {
		t.Errorf("expected version output, got:\n%s", stdout)
	}
}

func TestE2E_OnlyAndSkipMutuallyExclusive(t *testing.T) {
	_, stderr, code := runNamo("--only", "npm", "--skip", "domain", "test")

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got:\n%s", stderr)
	}
}

func TestE2E_UnknownFlag(t *testing.T) {
	_, stderr, code := runNamo("--nonexistent", "test")

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if stderr == "" {
		t.Error("expected error output for unknown flag")
	}
}

func TestE2E_NoRegistriesAfterFilter(t *testing.T) {
	_, stderr, code := runNamo("--only", "nonexistent", "test")

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, "no registries selected") {
		t.Errorf("expected 'no registries selected' error, got:\n%s", stderr)
	}
}

func TestE2E_OnlyFilter(t *testing.T) {
	stdout, _, _ := runNamo("--only", "npm", "--no-color", "react")

	if !strings.Contains(stdout, "npm") {
		t.Errorf("expected 'npm' in output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "GitHub") {
		t.Errorf("expected no GitHub in --only npm output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "Docker Hub") {
		t.Errorf("expected no Docker Hub in --only npm output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "Domain") {
		t.Errorf("expected no Domain in --only npm output, got:\n%s", stdout)
	}
}

func TestE2E_SkipFilter(t *testing.T) {
	stdout, _, _ := runNamo("--skip", "domain,dockerhub", "--no-color", "react")

	if !strings.Contains(stdout, "npm") {
		t.Errorf("expected 'npm' in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "GitHub") {
		t.Errorf("expected 'GitHub' in output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "Domain") {
		t.Errorf("expected no Domain in --skip domain output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "Docker Hub") {
		t.Errorf("expected no Docker Hub in --skip dockerhub output, got:\n%s", stdout)
	}
}

func TestE2E_NoColorFlag(t *testing.T) {
	stdout, _, _ := runNamo("--only", "npm", "--no-color", "react")

	if strings.Contains(stdout, "\033[") {
		t.Error("expected no ANSI escape codes with --no-color")
	}
}

func TestE2E_TakenName_ExitCode1(t *testing.T) {
	// "react" is taken on npm
	_, _, code := runNamo("--only", "npm", "react")

	if code != 1 {
		t.Errorf("expected exit code 1 for taken name, got %d", code)
	}
}

func TestE2E_TimeoutTriggersError(t *testing.T) {
	// 1ms timeout should cause all checks to fail
	stdout, _, code := runNamo("--timeout", "1ms", "--no-color", "test")

	// Should be exit 2 (errors) or possibly exit 1 if domain resolved from cache
	if code == 0 {
		t.Errorf("expected non-zero exit code with 1ms timeout, got 0.\nOutput:\n%s", stdout)
	}
}

func TestE2E_OutputContainsAvailableCount(t *testing.T) {
	stdout, _, _ := runNamo("--only", "npm", "--no-color", "react")

	if !strings.Contains(stdout, "of 1 available") {
		t.Errorf("expected 'of 1 available' in output, got:\n%s", stdout)
	}
}

func TestE2E_MultipleArgs(t *testing.T) {
	_, stderr, code := runNamo("name1", "name2")

	if code != 2 {
		t.Errorf("expected exit code 2 for multiple args, got %d", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected usage in stderr, got:\n%s", stderr)
	}
}
