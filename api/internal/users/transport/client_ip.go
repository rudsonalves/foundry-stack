package transport

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type ClientIPResolver struct {
	trustedProxies []netip.Prefix
}

func NewClientIPResolver(trustedProxyCIDRs []string) (*ClientIPResolver, error) {
	prefixes := make([]netip.Prefix, 0, len(trustedProxyCIDRs))

	for _, value := range trustedProxyCIDRs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("parse trusted proxy CIDR %q: %w", value, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}

	return &ClientIPResolver{trustedProxies: prefixes}, nil
}

func (r *ClientIPResolver) Resolve(request *http.Request) string {
	peer, ok := remoteAddress(request.RemoteAddr)
	if !ok {
		return "unknown"
	}

	if !r.isTrusted(peer) {
		return peer.String()
	}

	forwarded := request.Header.Values("X-Forwarded-For")
	if len(forwarded) == 0 {
		return peer.String()
	}

	chain := make([]netip.Addr, 0)
	for _, header := range forwarded {
		for _, value := range strings.Split(header, ",") {
			address, err := netip.ParseAddr(strings.TrimSpace(value))
			if err != nil {
				return peer.String()
			}
			chain = append(chain, address.Unmap())
		}
	}

	for index := len(chain) - 1; index >= 0; index-- {
		if !r.isTrusted(chain[index]) {
			return chain[index].String()
		}
	}

	if len(chain) > 0 {
		return chain[0].String()
	}

	return peer.String()
}

func (r *ClientIPResolver) isTrusted(address netip.Addr) bool {
	for _, prefix := range r.trustedProxies {
		if prefix.Contains(address) {
			return true
		}
	}

	return false
}

func remoteAddress(value string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(value)
	if err != nil {
		host = value
	}

	address, err := netip.ParseAddr(strings.TrimSpace(host))
	if err != nil {
		return netip.Addr{}, false
	}

	return address.Unmap(), true
}
