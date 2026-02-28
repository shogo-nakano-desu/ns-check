package runner

import (
	"context"
	"testing"
	"time"

	"github.com/shogonakano/ns-check/checker"
)

type mockChecker struct {
	name        string
	displayName string
	delay       time.Duration
	result      checker.Result
}

func (m *mockChecker) Name() string        { return m.name }
func (m *mockChecker) DisplayName() string { return m.displayName }
func (m *mockChecker) Check(ctx context.Context, name string) checker.Result {
	select {
	case <-time.After(m.delay):
		r := m.result
		r.Name = name
		return r
	case <-ctx.Done():
		return checker.Result{
			Registry: m.displayName,
			Name:     name,
			Status:   checker.Unknown,
			Err:      ctx.Err(),
		}
	}
}

func TestRun_AllComplete(t *testing.T) {
	checkers := []checker.Checker{
		&mockChecker{
			name:        "a",
			displayName: "Registry A",
			delay:       10 * time.Millisecond,
			result:      checker.Result{Registry: "Registry A", Status: checker.Available},
		},
		&mockChecker{
			name:        "b",
			displayName: "Registry B",
			delay:       10 * time.Millisecond,
			result:      checker.Result{Registry: "Registry B", Status: checker.Taken},
		},
		&mockChecker{
			name:        "c",
			displayName: "Registry C",
			delay:       10 * time.Millisecond,
			result:      checker.Result{Registry: "Registry C", Status: checker.Unknown, Err: context.DeadlineExceeded},
		},
	}

	results := Run(context.Background(), checkers, "testname")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Registry != "Registry A" || results[0].Status != checker.Available {
		t.Errorf("result[0]: got %+v", results[0])
	}
	if results[1].Registry != "Registry B" || results[1].Status != checker.Taken {
		t.Errorf("result[1]: got %+v", results[1])
	}
	if results[2].Registry != "Registry C" || results[2].Status != checker.Unknown {
		t.Errorf("result[2]: got %+v", results[2])
	}
	for _, r := range results {
		if r.Name != "testname" {
			t.Errorf("expected name 'testname', got %q", r.Name)
		}
	}
}

func TestRun_PreservesOrder(t *testing.T) {
	// Checker with longest delay is first — results should still be in original order.
	checkers := []checker.Checker{
		&mockChecker{
			name:        "slow",
			displayName: "Slow",
			delay:       100 * time.Millisecond,
			result:      checker.Result{Registry: "Slow", Status: checker.Available},
		},
		&mockChecker{
			name:        "fast",
			displayName: "Fast",
			delay:       1 * time.Millisecond,
			result:      checker.Result{Registry: "Fast", Status: checker.Taken},
		},
	}

	results := Run(context.Background(), checkers, "test")

	if results[0].Registry != "Slow" {
		t.Errorf("expected first result to be 'Slow', got %q", results[0].Registry)
	}
	if results[1].Registry != "Fast" {
		t.Errorf("expected second result to be 'Fast', got %q", results[1].Registry)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	checkers := []checker.Checker{
		&mockChecker{
			name:        "slow",
			displayName: "Slow",
			delay:       5 * time.Second,
			result:      checker.Result{Registry: "Slow", Status: checker.Available},
		},
	}

	results := Run(ctx, checkers, "test")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != checker.Unknown {
		t.Errorf("expected Unknown status on cancellation, got %v", results[0].Status)
	}
	if results[0].Err == nil {
		t.Error("expected error on cancelled result")
	}
}

func TestRun_ConcurrentExecution(t *testing.T) {
	// All checkers have 50ms delay. If sequential, total would be ~200ms.
	// If concurrent, total should be ~50ms.
	checkers := []checker.Checker{
		&mockChecker{name: "a", displayName: "A", delay: 50 * time.Millisecond, result: checker.Result{Registry: "A", Status: checker.Available}},
		&mockChecker{name: "b", displayName: "B", delay: 50 * time.Millisecond, result: checker.Result{Registry: "B", Status: checker.Available}},
		&mockChecker{name: "c", displayName: "C", delay: 50 * time.Millisecond, result: checker.Result{Registry: "C", Status: checker.Available}},
		&mockChecker{name: "d", displayName: "D", delay: 50 * time.Millisecond, result: checker.Result{Registry: "D", Status: checker.Available}},
	}

	start := time.Now()
	results := Run(context.Background(), checkers, "test")
	elapsed := time.Since(start)

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	// Should complete in well under 200ms if truly concurrent.
	if elapsed > 150*time.Millisecond {
		t.Errorf("expected concurrent execution (~50ms), but took %v", elapsed)
	}
}

func TestRun_EmptyCheckers(t *testing.T) {
	results := Run(context.Background(), nil, "test")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
