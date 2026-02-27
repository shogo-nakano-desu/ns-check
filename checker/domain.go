package checker

import (
	"context"
	"net"
	"strings"
)

// HostLookup abstracts DNS resolution for testability.
type HostLookup interface {
	LookupHost(ctx context.Context, host string) ([]string, error)
}

// DomainChecker checks .com domain availability via DNS lookup.
type DomainChecker struct {
	resolver HostLookup
}

func NewDomainChecker(resolver HostLookup) *DomainChecker {
	return &DomainChecker{resolver: resolver}
}

// NewDefaultDomainChecker creates a DomainChecker with the default net.Resolver.
func NewDefaultDomainChecker() *DomainChecker {
	return &DomainChecker{resolver: &net.Resolver{}}
}

func (c *DomainChecker) Name() string        { return "domain" }
func (c *DomainChecker) DisplayName() string { return "Domain (.com)" }

func (c *DomainChecker) Check(ctx context.Context, name string) Result {
	fqdn := name + ".com"

	addrs, err := c.resolver.LookupHost(ctx, fqdn)
	if err != nil {
		var dnsErr *net.DNSError
		if ok := isDNSNotFound(err, &dnsErr); ok {
			return Result{Registry: c.DisplayName(), Name: name, Status: Available}
		}
		return Result{Registry: c.DisplayName(), Name: name, Status: Unknown, Err: err}
	}

	if len(addrs) == 0 {
		return Result{Registry: c.DisplayName(), Name: name, Status: Available}
	}

	return Result{
		Registry: c.DisplayName(),
		Name:     name,
		Status:   Taken,
		Detail:   strings.Join(addrs, ", "),
	}
}

func isDNSNotFound(err error, dnsErr **net.DNSError) bool {
	if e, ok := err.(*net.DNSError); ok {
		*dnsErr = e
		return e.IsNotFound
	}
	return false
}
