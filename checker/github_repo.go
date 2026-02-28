package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GitHubRepoChecker checks if a repository with the given name exists on GitHub.
type GitHubRepoChecker struct {
	client  *http.Client
	baseURL string
	token   string
}

func NewGitHubRepoChecker(client *http.Client, baseURL string, token string) *GitHubRepoChecker {
	return &GitHubRepoChecker{client: client, baseURL: baseURL, token: token}
}

func (c *GitHubRepoChecker) Name() string        { return "github-repo" }
func (c *GitHubRepoChecker) DisplayName() string { return "GitHub Repo" }

func (c *GitHubRepoChecker) Check(ctx context.Context, name string) Result {
	u := c.baseURL + "/search/repositories?q=" + name + "+in:name&per_page=5"

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
		fullName := findExactRepoMatch(resp.Body, name)
		if fullName != "" {
			return Result{Registry: c.DisplayName(), Name: name, Status: Taken, Detail: fullName}
		}
		return Result{Registry: c.DisplayName(), Name: name, Status: Available}
	case http.StatusForbidden:
		return Result{
			Registry: c.DisplayName(),
			Name:     name,
			Status:   Unknown,
			Err:      fmt.Errorf("rate limited"),
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

func findExactRepoMatch(body io.Reader, name string) string {
	var data struct {
		Items []struct {
			Name     string `json:"name"`
			FullName string `json:"full_name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(io.LimitReader(body, 65536)).Decode(&data); err != nil {
		return ""
	}
	for _, item := range data.Items {
		if strings.EqualFold(item.Name, name) {
			return item.FullName
		}
	}
	return ""
}
