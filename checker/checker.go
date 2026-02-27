package checker

import "context"

// Status represents the result of a name availability check.
type Status int

const (
	Available Status = iota
	Taken
	Unknown
)

func (s Status) String() string {
	switch s {
	case Available:
		return "available"
	case Taken:
		return "taken"
	default:
		return "unknown"
	}
}

// Result holds the outcome of a single registry check.
type Result struct {
	Registry string
	Name     string
	Status   Status
	Err      error
	Detail   string
}

// Checker is the interface every registry checker must implement.
type Checker interface {
	// Name returns a short lowercase identifier (e.g. "npm") used for --only/--skip flags.
	Name() string

	// DisplayName returns a human-friendly label for output (e.g. "npm").
	DisplayName() string

	// Check tests whether the given name is available on this registry.
	Check(ctx context.Context, name string) Result
}
