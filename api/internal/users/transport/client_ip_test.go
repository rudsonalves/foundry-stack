package transport

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPResolverUsesDirectPeerByDefault(t *testing.T) {
	t.Parallel()

	resolver, err := NewClientIPResolver(nil)
	if err != nil {
		t.Fatalf("NewClientIPResolver() error = %v", err)
	}

	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "198.51.100.10:4321"
	request.Header.Set("X-Forwarded-For", "203.0.113.20")

	if got := resolver.Resolve(request); got != "198.51.100.10" {
		t.Fatalf("Resolve() = %q, want direct peer", got)
	}
}

func TestClientIPResolverAcceptsHeaderOnlyFromTrustedProxy(t *testing.T) {
	t.Parallel()

	resolver, err := NewClientIPResolver([]string{
		"10.0.0.0/8",
		"192.0.2.0/24",
	})
	if err != nil {
		t.Fatalf("NewClientIPResolver() error = %v", err)
	}

	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "10.0.0.5:4321"
	request.Header.Set(
		"X-Forwarded-For",
		"203.0.113.20, 192.0.2.15",
	)

	if got := resolver.Resolve(request); got != "203.0.113.20" {
		t.Fatalf("Resolve() = %q, want original untrusted client", got)
	}
}

func TestClientIPResolverFallsBackForMalformedForwardingChain(t *testing.T) {
	t.Parallel()

	resolver, err := NewClientIPResolver([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewClientIPResolver() error = %v", err)
	}

	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "10.0.0.5:4321"
	request.Header.Set("X-Forwarded-For", "not-an-ip")

	if got := resolver.Resolve(request); got != "10.0.0.5" {
		t.Fatalf("Resolve() = %q, want trusted peer fallback", got)
	}
}

func TestClientIPResolverHandlesIPv6AndUnknownPeer(t *testing.T) {
	t.Parallel()

	resolver, err := NewClientIPResolver(nil)
	if err != nil {
		t.Fatalf("NewClientIPResolver() error = %v", err)
	}

	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "[2001:db8::1]:4321"
	if got := resolver.Resolve(request); got != "2001:db8::1" {
		t.Fatalf("Resolve() = %q, want IPv6 peer", got)
	}

	request.RemoteAddr = ""
	if got := resolver.Resolve(request); got != "unknown" {
		t.Fatalf("Resolve() = %q, want unknown", got)
	}
}

func TestNewClientIPResolverRejectsInvalidCIDR(t *testing.T) {
	t.Parallel()

	if _, err := NewClientIPResolver([]string{"not-a-network"}); err == nil {
		t.Fatal("NewClientIPResolver() error = nil")
	}
}
