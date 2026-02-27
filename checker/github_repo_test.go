package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubRepoChecker_Taken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"total_count":1,"items":[{"name":"react","full_name":"facebook/react","stargazers_count":200000}]}`))
	}))
	defer srv.Close()

	c := NewGitHubRepoChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "react")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Detail != "facebook/react" {
		t.Errorf("expected detail 'facebook/react', got %q", result.Detail)
	}
}

func TestGitHubRepoChecker_Available_NoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"total_count":0,"items":[]}`))
	}))
	defer srv.Close()

	c := NewGitHubRepoChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "xyzzy-nonexistent")

	if result.Status != Available {
		t.Errorf("expected Available, got %v", result.Status)
	}
}

func TestGitHubRepoChecker_Available_NoExactMatch(t *testing.T) {
	// Search returns results, but none match the exact name.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"total_count":5,"items":[{"name":"react-native","full_name":"facebook/react-native","stargazers_count":100000}]}`))
	}))
	defer srv.Close()

	c := NewGitHubRepoChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "react")

	if result.Status != Available {
		t.Errorf("expected Available (no exact match), got %v", result.Status)
	}
}

func TestGitHubRepoChecker_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewGitHubRepoChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "test")

	if result.Status != Unknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
}

func TestGitHubRepoChecker_TokenSent(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"total_count":0,"items":[]}`))
	}))
	defer srv.Close()

	c := NewGitHubRepoChecker(srv.Client(), srv.URL, "ghp_testtoken")
	c.Check(context.Background(), "test")

	if receivedAuth != "Bearer ghp_testtoken" {
		t.Errorf("expected 'Bearer ghp_testtoken', got %q", receivedAuth)
	}
}

func TestGitHubRepoChecker_QueryParam(t *testing.T) {
	var receivedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"total_count":0,"items":[]}`))
	}))
	defer srv.Close()

	c := NewGitHubRepoChecker(srv.Client(), srv.URL, "")
	c.Check(context.Background(), "myproject")

	if receivedQuery != "myproject in:name" {
		t.Errorf("expected query 'myproject in:name', got %q", receivedQuery)
	}
}

func TestGitHubRepoChecker_Name(t *testing.T) {
	c := NewGitHubRepoChecker(http.DefaultClient, "", "")
	if c.Name() != "github-repo" {
		t.Errorf("expected name 'github-repo', got %q", c.Name())
	}
	if c.DisplayName() != "GitHub Repo" {
		t.Errorf("expected display name 'GitHub Repo', got %q", c.DisplayName())
	}
}

func TestGitHubRepoChecker_CaseInsensitiveMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"total_count":1,"items":[{"name":"React","full_name":"facebook/React","stargazers_count":200000}]}`))
	}))
	defer srv.Close()

	c := NewGitHubRepoChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "react")

	if result.Status != Taken {
		t.Errorf("expected Taken (case-insensitive match), got %v", result.Status)
	}
}
