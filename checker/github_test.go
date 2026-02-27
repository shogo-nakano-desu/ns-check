package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubChecker_Taken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"login":"octocat","type":"User"}`))
	}))
	defer srv.Close()

	c := NewGitHubChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "octocat")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Detail != "User" {
		t.Errorf("expected detail 'User', got %q", result.Detail)
	}
}

func TestGitHubChecker_TakenOrg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"login":"github","type":"Organization"}`))
	}))
	defer srv.Close()

	c := NewGitHubChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "github")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Detail != "Organization" {
		t.Errorf("expected detail 'Organization', got %q", result.Detail)
	}
}

func TestGitHubChecker_Available(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewGitHubChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "xyzzy-nonexistent")

	if result.Status != Available {
		t.Errorf("expected Available, got %v", result.Status)
	}
}

func TestGitHubChecker_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1700000000")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewGitHubChecker(srv.Client(), srv.URL, "")
	result := c.Check(context.Background(), "test")

	if result.Status != Unknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
	if result.Err == nil {
		t.Error("expected non-nil error for rate limit")
	}
}

func TestGitHubChecker_TokenSent(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewGitHubChecker(srv.Client(), srv.URL, "ghp_testtoken123")
	c.Check(context.Background(), "test")

	if receivedAuth != "Bearer ghp_testtoken123" {
		t.Errorf("expected 'Bearer ghp_testtoken123', got %q", receivedAuth)
	}
}

func TestGitHubChecker_NoTokenNoAuth(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewGitHubChecker(srv.Client(), srv.URL, "")
	c.Check(context.Background(), "test")

	if receivedAuth != "" {
		t.Errorf("expected no Authorization header, got %q", receivedAuth)
	}
}

func TestGitHubChecker_UserAgent(t *testing.T) {
	var receivedUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewGitHubChecker(srv.Client(), srv.URL, "")
	c.Check(context.Background(), "test")

	if receivedUA != "namo/1.0" {
		t.Errorf("expected User-Agent 'namo/1.0', got %q", receivedUA)
	}
}
