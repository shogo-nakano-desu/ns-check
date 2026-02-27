package checker

import (
	"context"
	"fmt"
	"net"
	"testing"
)

type fakeResolver struct {
	addrs []string
	err   error
}

func (f *fakeResolver) LookupHost(_ context.Context, _ string) ([]string, error) {
	return f.addrs, f.err
}

func TestDomainChecker_Taken(t *testing.T) {
	c := NewDomainChecker(&fakeResolver{addrs: []string{"93.184.216.34"}})
	result := c.Check(context.Background(), "example")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Detail != "93.184.216.34" {
		t.Errorf("expected detail '93.184.216.34', got %q", result.Detail)
	}
	if result.Registry != "Domain (.com)" {
		t.Errorf("expected registry 'Domain (.com)', got %q", result.Registry)
	}
}

func TestDomainChecker_TakenMultipleAddrs(t *testing.T) {
	c := NewDomainChecker(&fakeResolver{addrs: []string{"1.2.3.4", "5.6.7.8"}})
	result := c.Check(context.Background(), "example")

	if result.Status != Taken {
		t.Errorf("expected Taken, got %v", result.Status)
	}
	if result.Detail != "1.2.3.4, 5.6.7.8" {
		t.Errorf("expected detail '1.2.3.4, 5.6.7.8', got %q", result.Detail)
	}
}

func TestDomainChecker_Available(t *testing.T) {
	dnsErr := &net.DNSError{
		Err:        "no such host",
		Name:       "nonexistent.com",
		IsNotFound: true,
	}
	c := NewDomainChecker(&fakeResolver{err: dnsErr})
	result := c.Check(context.Background(), "nonexistent")

	if result.Status != Available {
		t.Errorf("expected Available, got %v", result.Status)
	}
}

func TestDomainChecker_UnknownError(t *testing.T) {
	c := NewDomainChecker(&fakeResolver{err: fmt.Errorf("network unreachable")})
	result := c.Check(context.Background(), "test")

	if result.Status != Unknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
	if result.Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestDomainChecker_DNSErrorNotFound(t *testing.T) {
	// A DNS error that is NOT "not found" (e.g., SERVFAIL) should be Unknown.
	dnsErr := &net.DNSError{
		Err:        "server failure",
		Name:       "test.com",
		IsNotFound: false,
	}
	c := NewDomainChecker(&fakeResolver{err: dnsErr})
	result := c.Check(context.Background(), "test")

	if result.Status != Unknown {
		t.Errorf("expected Unknown for non-NotFound DNS error, got %v", result.Status)
	}
}

func TestDomainChecker_AppendsDotCom(t *testing.T) {
	var queriedHost string
	resolver := &recordingResolver{inner: &fakeResolver{addrs: []string{"1.2.3.4"}}}
	c := NewDomainChecker(resolver)
	c.Check(context.Background(), "myproject")

	queriedHost = resolver.lastHost
	if queriedHost != "myproject.com" {
		t.Errorf("expected lookup for 'myproject.com', got %q", queriedHost)
	}
}

type recordingResolver struct {
	inner    HostLookup
	lastHost string
}

func (r *recordingResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	r.lastHost = host
	return r.inner.LookupHost(ctx, host)
}
