package ssrf_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"example.com/hookline/internal/ssrf"
)

// fakeDNS answers from a map, so the tests need no network.
func fakeDNS(m map[string][]string) ssrf.Lookup {
	return func(_ context.Context, host string) ([]netip.Addr, error) {
		var out []netip.Addr
		for _, s := range m[host] {
			out = append(out, netip.MustParseAddr(s))
		}
		if len(out) == 0 {
			return nil, errors.New("no such host")
		}
		return out, nil
	}
}

var dns = fakeDNS(map[string][]string{
	"hooks.example.com":   {"93.184.215.14"},
	"internal.example":    {"10.1.2.3"},
	"mixed.example":       {"93.184.215.14", "127.0.0.1"}, // one bad answer is enough to refuse
	"sink":                {"172.18.0.5"},                 // a compose service
	"sink.hookline.svc":   {"10.96.12.34"},                // a kind Service
	"metadata.example":    {"169.254.169.254"},
	"v6-loopback.example": {"::1"},
})

func TestCheckResolved(t *testing.T) {
	tests := []struct {
		url     string
		dev     bool
		allowed bool
		err     error
	}{
		{"https://hooks.example.com/x", false, true, nil},
		{"https://internal.example/x", false, false, ssrf.ErrBlocked},
		{"https://mixed.example/x", false, false, ssrf.ErrBlocked},
		{"https://metadata.example/latest", true, false, ssrf.ErrBlocked}, // not even in dev mode
		{"https://nowhere.example/x", false, false, ssrf.ErrUnresolvable},
		{"http://hooks.example.com/x", false, false, ssrf.ErrBlocked}, // http only in dev mode
		{"https://10.0.0.1/x", false, false, ssrf.ErrBlocked},         // a literal, checked without DNS
		{"http://sink:8081/hooks/a", true, true, nil},
		{"http://sink.hookline.svc:8081/hooks/a", true, true, nil},
		{"http://localhost:8091/hooks/a", true, false, ssrf.ErrUnresolvable}, // the fake DNS has no localhost
		{"http://v6-loopback.example/x", true, true, nil},
		{"http://sink:8081/hooks/a", false, false, ssrf.ErrBlocked},
	}
	for _, tt := range tests {
		err := ssrf.ForMode(tt.dev).CheckResolved(context.Background(), dns, tt.url)
		if (err == nil) != tt.allowed || (tt.err != nil && !errors.Is(err, tt.err)) {
			t.Errorf("CheckResolved(%s, dev=%v) = %v, want allowed=%v (%v)", tt.url, tt.dev, err, tt.allowed, tt.err)
		}
	}
}

func TestDevAllowStaysNarrow(t *testing.T) {
	dev := ssrf.ForMode(true)
	for _, ip := range []string{"169.254.169.254", "10.0.0.1", "100.64.0.1", "fd00:ec2::254"} {
		if dev.CheckAddr(netip.MustParseAddr(ip)) == nil {
			t.Errorf("dev mode allows %s; only loopback, Docker and kind ranges may be allowed", ip)
		}
	}
}
