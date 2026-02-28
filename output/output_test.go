package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/shogonakano/ns-check/checker"
)

func TestPrinter_Available(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, false)
	results := []checker.Result{
		{Registry: "npm", Name: "test", Status: checker.Available},
	}
	p.Print("test", results)

	out := buf.String()
	if !strings.Contains(out, "✓ available") {
		t.Errorf("expected '✓ available' in output, got:\n%s", out)
	}
}

func TestPrinter_Taken(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, false)
	results := []checker.Result{
		{Registry: "npm", Name: "test", Status: checker.Taken},
	}
	p.Print("test", results)

	out := buf.String()
	if !strings.Contains(out, "✗ taken") {
		t.Errorf("expected '✗ taken' in output, got:\n%s", out)
	}
}

func TestPrinter_Error(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, false)
	results := []checker.Result{
		{Registry: "npm", Name: "test", Status: checker.Unknown, Err: fmt.Errorf("connection timeout")},
	}
	p.Print("test", results)

	out := buf.String()
	if !strings.Contains(out, "⚠") {
		t.Errorf("expected '⚠' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "connection timeout") {
		t.Errorf("expected error message in output, got:\n%s", out)
	}
}

func TestPrinter_Summary(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, false)
	results := []checker.Result{
		{Registry: "npm", Name: "test", Status: checker.Available},
		{Registry: "GitHub", Name: "test", Status: checker.Taken},
		{Registry: "Docker Hub", Name: "test", Status: checker.Available},
	}
	p.Print("test", results)

	out := buf.String()
	if !strings.Contains(out, "2 of 3 available") {
		t.Errorf("expected '2 of 3 available' in output, got:\n%s", out)
	}
}

func TestPrinter_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, false)
	results := []checker.Result{
		{Registry: "npm", Name: "test", Status: checker.Available},
	}
	p.Print("test", results)

	out := buf.String()
	if strings.Contains(out, "\033[") {
		t.Errorf("expected no ANSI codes in no-color mode, got:\n%s", out)
	}
}

func TestPrinter_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, true)
	results := []checker.Result{
		{Registry: "npm", Name: "test", Status: checker.Available},
	}
	p.Print("test", results)

	out := buf.String()
	if !strings.Contains(out, "\033[") {
		t.Errorf("expected ANSI codes in color mode, got:\n%s", out)
	}
}

func TestPrinter_Detail(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, false)
	results := []checker.Result{
		{Registry: "Domain (.com)", Name: "test", Status: checker.Taken, Detail: "93.184.216.34"},
	}
	p.Print("test", results)

	out := buf.String()
	if !strings.Contains(out, "93.184.216.34") {
		t.Errorf("expected detail '93.184.216.34' in output, got:\n%s", out)
	}
}

func TestPrinter_NameInHeader(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, false)
	results := []checker.Result{
		{Registry: "npm", Name: "myproject", Status: checker.Available},
	}
	p.Print("myproject", results)

	out := buf.String()
	if !strings.Contains(out, "myproject") {
		t.Errorf("expected name 'myproject' in header, got:\n%s", out)
	}
}

func TestPrinter_MultipleRegistries_Alignment(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinterWithWriter(&buf, false)
	results := []checker.Result{
		{Registry: "npm", Name: "test", Status: checker.Available},
		{Registry: "Domain (.com)", Name: "test", Status: checker.Taken},
	}
	p.Print("test", results)

	out := buf.String()
	// Both registry names should appear
	if !strings.Contains(out, "npm") {
		t.Errorf("expected 'npm' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Domain (.com)") {
		t.Errorf("expected 'Domain (.com)' in output, got:\n%s", out)
	}
}
