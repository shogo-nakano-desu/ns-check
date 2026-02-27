package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCratesChecker_Taken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"crate":{"name":"serde"}}`))
	}))
	defer srv.Close()

	c := NewCratesChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "serde")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Registry != "crates.io" {
		t.Errorf("expected registry 'crates.io', got %q", result.Registry)
	}
}

func TestCratesChecker_Available(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewCratesChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "xyzzy-nonexistent")

	if result.Status != Available {
		t.Errorf("expected Available, got %v", result.Status)
	}
}

func TestCratesChecker_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewCratesChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "test")

	if result.Status != Unknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
	if result.Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestCratesChecker_URLPath(t *testing.T) {
	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewCratesChecker(srv.Client(), srv.URL)
	c.Check(context.Background(), "my-crate")

	if receivedPath != "/api/v1/crates/my-crate" {
		t.Errorf("expected path '/api/v1/crates/my-crate', got %q", receivedPath)
	}
}

func TestCratesChecker_UserAgent(t *testing.T) {
	var receivedUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewCratesChecker(srv.Client(), srv.URL)
	c.Check(context.Background(), "test")

	if receivedUA != "namo/1.0" {
		t.Errorf("expected User-Agent 'namo/1.0', got %q", receivedUA)
	}
}

func TestCratesChecker_Name(t *testing.T) {
	c := NewCratesChecker(http.DefaultClient, "")
	if c.Name() != "crates" {
		t.Errorf("expected name 'crates', got %q", c.Name())
	}
}
