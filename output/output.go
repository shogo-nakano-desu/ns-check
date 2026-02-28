package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shogonakano/ns-check/checker"
)

const (
	reset  = "\033[0m"
	green  = "\033[32m"
	red    = "\033[31m"
	yellow = "\033[33m"
	bold   = "\033[1m"
	dim    = "\033[2m"
)

// Printer formats and prints check results to a writer.
type Printer struct {
	w     io.Writer
	color bool
}

// NewPrinter creates a Printer that writes to stdout with the given color setting.
func NewPrinter(color bool) *Printer {
	return &Printer{w: os.Stdout, color: color}
}

// NewPrinterWithWriter creates a Printer with a custom writer (useful for testing).
func NewPrinterWithWriter(w io.Writer, color bool) *Printer {
	return &Printer{w: w, color: color}
}

// Print formats and outputs the check results.
func (p *Printer) Print(name string, results []checker.Result) {
	_, _ = fmt.Fprintf(p.w, "\nChecking availability for %s\n\n", p.styled(bold, `"`+name+`"`))

	maxLen := 0
	for _, r := range results {
		if len(r.Registry) > maxLen {
			maxLen = len(r.Registry)
		}
	}

	for _, r := range results {
		padding := strings.Repeat(" ", maxLen-len(r.Registry)+2)
		switch r.Status {
		case checker.Available:
			_, _ = fmt.Fprintf(p.w, "  %s%s%s\n", r.Registry, padding, p.styled(green, "✓ available"))
		case checker.Taken:
			_, _ = fmt.Fprintf(p.w, "  %s%s%s\n", r.Registry, padding, p.styled(red, "✗ taken"))
		case checker.Unknown:
			errMsg := "unknown error"
			if r.Err != nil {
				errMsg = r.Err.Error()
			}
			_, _ = fmt.Fprintf(p.w, "  %s%s%s\n", r.Registry, padding, p.styled(yellow, "⚠ "+errMsg))
		}
		if r.Detail != "" {
			detailPad := strings.Repeat(" ", maxLen+4)
			_, _ = fmt.Fprintf(p.w, "%s%s\n", detailPad, p.styled(dim, r.Detail))
		}
	}

	avail := 0
	for _, r := range results {
		if r.Status == checker.Available {
			avail++
		}
	}
	_, _ = fmt.Fprintf(p.w, "\n%d of %d available\n", avail, len(results))
}

func (p *Printer) styled(code, text string) string {
	if !p.color {
		return text
	}
	return code + text + reset
}

// IsTerminal returns true if stdout is a terminal (not piped).
func IsTerminal() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// ShouldUseColor returns true if color output should be enabled.
func ShouldUseColor(noColorFlag bool) bool {
	if noColorFlag {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return IsTerminal()
}
