package ssrf

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
)

// DevAllow are the only private ranges Hookline's dev mode may reach (milestone M9): loopback for the
// receivers on your machine, Docker's default network ranges for compose services, and kind's default
// pod (10.244.0.0/16) and service (10.96.0.0/12) ranges for the in-cluster sink. Production allows none.
var DevAllow = []netip.Prefix{
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("10.96.0.0/12"),
	netip.MustParsePrefix("10.244.0.0/16"),
}

// ForMode is Hookline's policy: https to public addresses only, or, in dev mode, http too and DevAllow.
func ForMode(dev bool) Policy {
	if dev {
		return Policy{Schemes: []string{"http", "https"}, Allow: DevAllow}
	}
	return Policy{Schemes: []string{"https"}}
}

// Lookup resolves a host name. LookupIP is the real one; tests pass a fake.
type Lookup func(ctx context.Context, host string) ([]netip.Addr, error)

// LookupIP resolves with the system resolver, IPv4 and IPv6 addresses both.
func LookupIP(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
}

// ErrUnresolvable means the host name has no address at all.
var ErrUnresolvable = errors.New("ssrf: host does not resolve")

// CheckResolved is the check at endpoint creation: CheckURL, then every address the host resolves to
// must be allowed. It catches the honest mistake and the obvious attack early, with a clear answer;
// the dialer's Control at send time remains the check that holds (DNS can change in between).
func (p Policy) CheckResolved(ctx context.Context, lookup Lookup, raw string) error {
	u, err := p.CheckURL(raw)
	if err != nil {
		return err
	}
	if _, err := netip.ParseAddr(u.Hostname()); err == nil {
		return nil // a literal address: CheckURL has checked it
	}
	addrs, err := lookup(ctx, u.Hostname())
	if err != nil || len(addrs) == 0 {
		return fmt.Errorf("%w: %s", ErrUnresolvable, u.Hostname())
	}
	for _, a := range addrs {
		if err := p.CheckAddr(a); err != nil {
			return err
		}
	}
	return nil
}
