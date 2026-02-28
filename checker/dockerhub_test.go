package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDockerHubChecker_Taken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"library","username":"library"}`))
	}))
	defer srv.Close()

	c := NewDockerHubChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "library")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Registry != "Docker Hub" {
		t.Errorf("expected registry 'Docker Hub', got %q", result.Registry)
	}
}

func TestDockerHubChecker_Available(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewDockerHubChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "xyzzy-nonexistent")

	if result.Status != Available {
		t.Errorf("expected Available, got %v", result.Status)
	}
}

func TestDockerHubChecker_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := NewDockerHubChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "test")

	if result.Status != Unknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
	if result.Err == nil {
		t.Error("expected non-nil error for rate limit")
	}
}

func TestDockerHubChecker_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewDockerHubChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "test")

	if result.Status != Unknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
}

func TestDockerHubChecker_URLPath(t *testing.T) {
	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewDockerHubChecker(srv.Client(), srv.URL)
	c.Check(context.Background(), "myproject")

	if receivedPath != "/v2/users/myproject" {
		t.Errorf("expected path '/v2/users/myproject', got %q", receivedPath)
	}
}

func TestDockerHubChecker_UserAgent(t *testing.T) {
	var receivedUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewDockerHubChecker(srv.Client(), srv.URL)
	c.Check(context.Background(), "test")

	if receivedUA != "ns-check/1.0" {
		t.Errorf("expected User-Agent 'ns-check/1.0', got %q", receivedUA)
	}
}
