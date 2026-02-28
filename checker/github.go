package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GitHubChecker checks username/organization availability on GitHub.
type GitHubChecker struct {
	client  *http.Client
	baseURL string
	token   string
}

func NewGitHubChecker(client *http.Client, baseURL string, token string) *GitHubChecker {
	return &GitHubChecker{client: client, baseURL: baseURL, token: token}
}

func (c *GitHubChecker) Name() string        { return "github" }
func (c *GitHubChecker) DisplayName() string { return "GitHub" }

func (c *GitHubChecker) Check(ctx context.Context, name string) Result {
	u := c.baseURL + "/users/" + url.PathEscape(name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Result{Registry: c.DisplayName(), Name: name, Status: Unknown, Err: err}
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ns-check/1.0")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{Registry: c.DisplayName(), Name: name, Status: Unknown, Err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		detail := parseGitHubType(resp.Body)
		return Result{Registry: c.DisplayName(), Name: name, Status: Taken, Detail: detail}
	case http.StatusNotFound:
		return Result{Registry: c.DisplayName(), Name: name, Status: Available}
	case http.StatusForbidden:
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return Result{
				Registry: c.DisplayName(),
				Name:     name,
				Status:   Unknown,
				Err:      fmt.Errorf("rate limited"),
			}
		}
		return Result{
			Registry: c.DisplayName(),
			Name:     name,
			Status:   Unknown,
			Err:      fmt.Errorf("forbidden (status 403)"),
		}
	default:
		return Result{
			Registry: c.DisplayName(),
			Name:     name,
			Status:   Unknown,
			Err:      fmt.Errorf("unexpected status: %d", resp.StatusCode),
		}
	}
}

func parseGitHubType(body io.Reader) string {
	var data struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(io.LimitReader(body, 4096)).Decode(&data); err != nil {
		return ""
	}
	return data.Type
}
