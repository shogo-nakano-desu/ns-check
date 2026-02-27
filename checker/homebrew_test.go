package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomebrewChecker_TakenFormula(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/formula/wget.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"wget"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewHomebrewChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "wget")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Detail != "formula" {
		t.Errorf("expected detail 'formula', got %q", result.Detail)
	}
}

func TestHomebrewChecker_TakenCask(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/cask/firefox.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"token":"firefox"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewHomebrewChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "firefox")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Detail != "cask" {
		t.Errorf("expected detail 'cask', got %q", result.Detail)
	}
}

func TestHomebrewChecker_TakenBoth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Both formula and cask exist
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewHomebrewChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "both")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Detail != "formula, cask" {
		t.Errorf("expected detail 'formula, cask', got %q", result.Detail)
	}
}

func TestHomebrewChecker_Available(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewHomebrewChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "xyzzy-nonexistent")

	if result.Status != Available {
		t.Errorf("expected Available, got %v", result.Status)
	}
}

func TestHomebrewChecker_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewHomebrewChecker(srv.Client(), srv.URL)
	result := c.Check(context.Background(), "test")

	if result.Status != Unknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
}

func TestHomebrewChecker_Name(t *testing.T) {
	c := NewHomebrewChecker(http.DefaultClient, "")
	if c.Name() != "homebrew" {
		t.Errorf("expected name 'homebrew', got %q", c.Name())
	}
	if c.DisplayName() != "Homebrew" {
		t.Errorf("expected display name 'Homebrew', got %q", c.DisplayName())
	}
}

func TestHomebrewChecker_URLPaths(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewHomebrewChecker(srv.Client(), srv.URL)
	c.Check(context.Background(), "myproject")

	expectedFormula := "/api/formula/myproject.json"
	expectedCask := "/api/cask/myproject.json"

	foundFormula, foundCask := false, false
	for _, p := range paths {
		if p == expectedFormula {
			foundFormula = true
		}
		if p == expectedCask {
			foundCask = true
		}
	}
	if !foundFormula {
		t.Errorf("expected formula path %q in requests, got %v", expectedFormula, paths)
	}
	if !foundCask {
		t.Errorf("expected cask path %q in requests, got %v", expectedCask, paths)
	}
}
