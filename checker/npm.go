package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// NpmChecker checks package name availability on the npm registry.
type NpmChecker struct {
	client  *http.Client
	baseURL string
}

func NewNpmChecker(client *http.Client, baseURL string) *NpmChecker {
	return &NpmChecker{client: client, baseURL: baseURL}
}

func (c *NpmChecker) Name() string        { return "npm" }
func (c *NpmChecker) DisplayName() string { return "npm" }

func (c *NpmChecker) Check(ctx context.Context, name string) Result {
	u := c.baseURL + "/" + url.PathEscape(name)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
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
