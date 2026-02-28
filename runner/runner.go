package runner

import (
	"context"

	"github.com/shogonakano/ns-check/checker"
)

// Run executes all checkers concurrently and returns results in the same order as input.
func Run(ctx context.Context, checkers []checker.Checker, name string) []checker.Result {
	if len(checkers) == 0 {
		return nil
	}

	type indexedResult struct {
		index  int
		result checker.Result
	}

	ch := make(chan indexedResult, len(checkers))

	for i, c := range checkers {
		go func(idx int, chk checker.Checker) {
			defer func() {
				if r := recover(); r != nil {
					ch <- indexedResult{
						index: idx,
						result: checker.Result{
							Registry: chk.DisplayName(),
							Name:     name,
							Status:   checker.Unknown,
							Err:      context.Canceled,
						},
					}
				}
			}()
			ch <- indexedResult{index: idx, result: chk.Check(ctx, name)}
		}(i, c)
	}

	results := make([]checker.Result, len(checkers))
	for range checkers {
		ir := <-ch
		results[ir.index] = ir.result
	}

	return results
}
