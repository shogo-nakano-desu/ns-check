package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNpmChecker_Taken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewNpmChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "express")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Registry != "npm" {
		t.Errorf("expected registry 'npm', got %q", result.Registry)
	}
	if result.Name != "express" {
		t.Errorf("expected name 'express', got %q", result.Name)
	}
}

func TestNpmChecker_Available(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewNpmChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "nonexistent-pkg-xyz")

	if result.Status != Available {
		t.Errorf("expected Available, got %v", result.Status)
	}
}

func TestNpmChecker_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewNpmChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "some-pkg")

	if result.Status != Unknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
	if result.Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestNpmChecker_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	c := NewNpmChecker(srv.Client(), srv.URL)
	result := c.Check(ctx, "anything")

	if result.Status != Unknown {
		t.Errorf("expected Unknown on timeout, got %v", result.Status)
	}
	if result.Err == nil {
		t.Error("expected non-nil error on timeout")
	}
}

func TestNpmChecker_HeadMethod(t *testing.T) {
	var receivedMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewNpmChecker(srv.Client(), srv.URL)
	c.Check(context.Background(), "test")

	if receivedMethod != http.MethodHead {
		t.Errorf("expected HEAD method, got %q", receivedMethod)
	}
}

func TestNpmChecker_UserAgent(t *testing.T) {
	var receivedUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewNpmChecker(srv.Client(), srv.URL)
	c.Check(context.Background(), "test")

	if receivedUA != "nsprobe/1.0" {
		t.Errorf("expected User-Agent 'nsprobe/1.0', got %q", receivedUA)
	}
}
