package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// DockerHubChecker checks namespace availability on Docker Hub.
type DockerHubChecker struct {
	client  *http.Client
	baseURL string
}

func NewDockerHubChecker(client *http.Client, baseURL string) *DockerHubChecker {
	return &DockerHubChecker{client: client, baseURL: baseURL}
}

func (c *DockerHubChecker) Name() string        { return "dockerhub" }
func (c *DockerHubChecker) DisplayName() string { return "Docker Hub" }

func (c *DockerHubChecker) Check(ctx context.Context, name string) Result {
	u := c.baseURL + "/v2/users/" + url.PathEscape(name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Result{Registry: c.DisplayName(), Name: name, Status: Unknown, Err: err}
	}
	req.Header.Set("User-Agent", "nsprobe/1.0")

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
	case http.StatusTooManyRequests:
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
