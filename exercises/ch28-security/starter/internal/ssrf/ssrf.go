// Package ssrf makes outbound HTTP requests to user-supplied URLs without
// letting them reach loopback, private, link-local or cloud-metadata addresses.
//
// Two layers: CheckURL rejects bad URLs early with a clear error, and the
// dialer's Control hook checks the address actually being connected to,
// after DNS resolution. The second layer is the one that holds.
package ssrf

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"syscall"
	"time"
)

// ErrBlocked is wrapped by every refusal of this package.
var ErrBlocked = errors.New("ssrf: destination not allowed")

// blocked lists the ranges no user-supplied URL may reach.
var blocked = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),      // "this network"; 0.0.0.0 reaches localhost on Linux
	netip.MustParsePrefix("10.0.0.0/8"),     // private
	netip.MustParsePrefix("100.64.0.0/10"),  // carrier-grade NAT, often internal
	netip.MustParsePrefix("127.0.0.0/8"),    // loopback
	netip.MustParsePrefix("169.254.0.0/16"), // link-local, incl. 169.254.169.254 metadata
	netip.MustParsePrefix("172.16.0.0/12"),  // private
	netip.MustParsePrefix("192.0.0.0/24"),   // IETF protocol assignments
	netip.MustParsePrefix("192.168.0.0/16"), // private
	netip.MustParsePrefix("198.18.0.0/15"),  // benchmarking
	netip.MustParsePrefix("224.0.0.0/4"),    // multicast
	netip.MustParsePrefix("240.0.0.0/4"),    // reserved, incl. broadcast
	netip.MustParsePrefix("::/128"),         // unspecified
	netip.MustParsePrefix("::1/128"),        // loopback
	netip.MustParsePrefix("64:ff9b::/96"),   // NAT64: embeds an IPv4 address
	netip.MustParsePrefix("fc00::/7"),       // unique local, incl. fd00:ec2::254 metadata
	netip.MustParsePrefix("fe80::/10"),      // link-local
	netip.MustParsePrefix("ff00::/8"),       // multicast
	netip.MustParsePrefix("2001:db8::/32"),  // documentation
	netip.MustParsePrefix("100::/64"),       // discard-only
	netip.MustParsePrefix("2002::/16"),      // 6to4: embeds an IPv4 address
	netip.MustParsePrefix("2001::/32"),      // Teredo: embeds an IPv4 address
}

// Policy decides which destinations are allowed.
type Policy struct {
	// Schemes allowed in URLs; empty means https only.
	Schemes []string
	// Allow lists exceptions to the blocked ranges, for development only
	// (for example 127.0.0.1/32 to reach a local receiver).
	Allow []netip.Prefix
}

func (p Policy) schemes() []string {
	if len(p.Schemes) == 0 {
		return []string{"https"}
	}
	return p.Schemes
}

// CheckAddr reports whether ip may be connected to.
func (p Policy) CheckAddr(ip netip.Addr) error {
	// TODO(day 84): unmap IPv4-mapped IPv6 first, return nil for an address in
	// p.Allow, refuse an invalid address, and refuse any address inside one of
	// the blocked prefixes with an error that wraps ErrBlocked.
	panic("TODO(day 84): CheckAddr")
}

// CheckURL parses raw and rejects what can be rejected from the string alone:
// the scheme, credentials in the URL, a missing host, and literal IP addresses.
// A hostname passes here; the dialer checks what it resolves to.
func (p Policy) CheckURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBlocked, err)
	}
	if !slices.Contains(p.schemes(), u.Scheme) {
		return nil, fmt.Errorf("%w: scheme %q", ErrBlocked, u.Scheme)
	}
	if u.User != nil {
		return nil, fmt.Errorf("%w: credentials in URL", ErrBlocked)
	}
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("%w: no host", ErrBlocked)
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if err := p.CheckAddr(ip); err != nil {
			return nil, err
		}
	}
	return u, nil
}

// Control is a net.Dialer Control hook. It runs after DNS resolution and
// before connect(2), with the exact address of this connection attempt, so a
// hostname that resolves to 127.0.0.1 (or switches to it: DNS rebinding) is
// refused here even though CheckURL accepted the name.
func (p Policy) Control(network, address string, _ syscall.RawConn) error {
	// TODO(day 84): allow only tcp4 and tcp6, parse address with
	// netip.ParseAddrPort, and return p.CheckAddr of its address.
	panic("TODO(day 84): Control")
}

// maxRedirects bounds the redirect chain; each hop is re-checked.
const maxRedirects = 5

// Client returns an *http.Client whose every connection passes Control and
// whose every redirect target passes CheckURL.
func (p Policy) Client(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, Control: p.Control}
	transport := &http.Transport{
		Proxy:                 nil, // a proxy would make the dialer check the proxy, not the target
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// TODO(day 84): stop after maxRedirects hops (wrap ErrBlocked) and
			// run every redirect target through p.CheckURL.
			return nil
		},
	}
}
