package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// HomebrewChecker checks formula/cask name availability on Homebrew.
type HomebrewChecker struct {
	client  *http.Client
	baseURL string
}

func NewHomebrewChecker(client *http.Client, baseURL string) *HomebrewChecker {
	return &HomebrewChecker{client: client, baseURL: baseURL}
}

func (c *HomebrewChecker) Name() string        { return "homebrew" }
func (c *HomebrewChecker) DisplayName() string { return "Homebrew" }

func (c *HomebrewChecker) Check(ctx context.Context, name string) Result {
	formulaExists, formulaErr := c.checkEndpoint(ctx, "/api/formula/"+url.PathEscape(name)+".json")
	caskExists, caskErr := c.checkEndpoint(ctx, "/api/cask/"+url.PathEscape(name)+".json")

	// If both errored, report unknown.
	if formulaErr != nil && caskErr != nil {
		return Result{
			Registry: c.DisplayName(),
			Name:     name,
			Status:   Unknown,
			Err:      fmt.Errorf("formula: %v; cask: %v", formulaErr, caskErr),
		}
	}

	var found []string
	if formulaExists {
		found = append(found, "formula")
	}
	if caskExists {
		found = append(found, "cask")
	}

	if len(found) > 0 {
		return Result{
			Registry: c.DisplayName(),
			Name:     name,
			Status:   Taken,
			Detail:   strings.Join(found, ", "),
		}
	}

	return Result{Registry: c.DisplayName(), Name: name, Status: Available}
}

// checkEndpoint returns (exists, error).
func (c *HomebrewChecker) checkEndpoint(ctx context.Context, path string) (bool, error) {
	u := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "nsprobe/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}
