package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CratesChecker checks crate name availability on crates.io.
type CratesChecker struct {
	client  *http.Client
	baseURL string
}

func NewCratesChecker(client *http.Client, baseURL string) *CratesChecker {
	return &CratesChecker{client: client, baseURL: baseURL}
}

func (c *CratesChecker) Name() string        { return "crates" }
func (c *CratesChecker) DisplayName() string { return "crates.io" }

func (c *CratesChecker) Check(ctx context.Context, name string) Result {
	u := c.baseURL + "/api/v1/crates/" + url.PathEscape(name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Result{Registry: c.DisplayName(), Name: name, Status: Unknown, Err: err}
	}
	req.Header.Set("User-Agent", "ns-check/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{Registry: c.DisplayName(), Name: name, Status: Unknown, Err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return Result{Registry: c.DisplayName(), Name: name, Status: Taken}
	case http.StatusNotFound:
		return Result{Registry: c.DisplayName(), Name: name, Status: Available}
	default:
		return Result{
			Registry: c.DisplayName(),
			Name:     name,
			Status:   Unknown,
			Err:      fmt.Errorf("unexpected status: %d", resp.StatusCode),
		}
	}
}
