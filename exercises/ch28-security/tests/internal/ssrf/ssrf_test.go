package ssrf_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"example.com/ch28-security/internal/ssrf"
)

func TestCheckAddr(t *testing.T) {
	tests := []struct {
		addr    string
		blocked bool
	}{
		{"93.184.215.14", false},
		{"2606:4700:4700::1111", false},
		{"127.0.0.1", true},
		{"127.10.20.30", true},
		{"::1", true},
		{"::ffff:127.0.0.1", true}, // IPv4-mapped loopback
		{"10.1.2.3", true},
		{"172.31.255.255", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true}, // cloud metadata (AWS, GCP, Azure)
		{"fd00:ec2::254", true},   // AWS metadata over IPv6
		{"fe80::1", true},
		{"100.64.0.1", true},
		{"0.0.0.0", true},
		{"224.0.0.1", true},
		{"64:ff9b::7f00:1", true}, // NAT64 form of 127.0.0.1
	}
	var p ssrf.Policy
	for _, tc := range tests {
		t.Run(tc.addr, func(t *testing.T) {
			err := p.CheckAddr(netip.MustParseAddr(tc.addr))
			if got := err != nil; got != tc.blocked {
				t.Fatalf("CheckAddr(%s) = %v; want blocked=%v", tc.addr, err, tc.blocked)
			}
			if err != nil && !errors.Is(err, ssrf.ErrBlocked) {
				t.Fatalf("error %v does not wrap ErrBlocked", err)
			}
		})
	}
}

func TestCheckURL(t *testing.T) {
	tests := []struct {
		raw string
		ok  bool
	}{
		{"https://hooks.example.com/in", true},
		{"https://hooks.example.com:8443/in?x=1", true},
		{"http://hooks.example.com/in", false}, // https only by default
		{"ftp://hooks.example.com/", false},
		{"file:///etc/passwd", false},
		{"https://user:pass@hooks.example.com/", false},
		{"https:///no-host", false},
		{"https://127.0.0.1/", false},
		{"https://[::1]:8080/", false},
		{"https://169.254.169.254/latest/meta-data/", false},
		{"https://[::ffff:10.0.0.1]/", false},
		{"https://localhost/", true}, // a name: only the dialer can tell where it goes
	}
	var p ssrf.Policy
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			_, err := p.CheckURL(tc.raw)
			if got := err == nil; got != tc.ok {
				t.Fatalf("CheckURL(%q) = %v; want ok=%v", tc.raw, err, tc.ok)
			}
		})
	}
}

// countingServer is a local HTTP server that counts the requests it receives.
func countingServer(t *testing.T, h http.HandlerFunc) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func ok(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }

var httpPolicy = ssrf.Policy{Schemes: []string{"http", "https"}}

func TestClientRefusesLoopbackServer(t *testing.T) {
	srv, hits := countingServer(t, ok)
	_, err := httpPolicy.Client(2 * time.Second).Get(srv.URL) // http://127.0.0.1:port
	if !errors.Is(err, ssrf.ErrBlocked) {
		t.Fatalf("Get(%s) error = %v; want ErrBlocked", srv.URL, err)
	}
	if hits.Load() != 0 {
		t.Fatalf("server saw %d requests; the connection must never be made", hits.Load())
	}
}

// TestClientChecksResolvedAddress is the DNS-rebinding case: the URL names a
// host, CheckURL accepts it, and only the dialer sees that it resolves to loopback.
func TestClientChecksResolvedAddress(t *testing.T) {
	srv, hits := countingServer(t, ok)
	u, _ := url.Parse(srv.URL)
	byName := "http://localhost:" + u.Port() + "/"
	if _, err := httpPolicy.CheckURL(byName); err != nil {
		t.Fatalf("CheckURL(%q) = %v; a hostname should pass the string check", byName, err)
	}
	_, err := httpPolicy.Client(2 * time.Second).Get(byName)
	if !errors.Is(err, ssrf.ErrBlocked) {
		t.Fatalf("Get(%s) error = %v; want ErrBlocked from the dialer", byName, err)
	}
	if hits.Load() != 0 {
		t.Fatalf("server saw %d requests", hits.Load())
	}
}

// TestClientRefusesOddIPSpelling: "0x7f000001" is not an IP literal for
// url/netip, so CheckURL lets it through as a name. Some resolvers turn it
// into 127.0.0.1, others fail; either way no request may arrive.
func TestClientRefusesOddIPSpelling(t *testing.T) {
	srv, hits := countingServer(t, ok)
	u, _ := url.Parse(srv.URL)
	_, err := httpPolicy.Client(2 * time.Second).Get("http://0x7f000001:" + u.Port() + "/")
	if err == nil || hits.Load() != 0 {
		t.Fatalf("err = %v, hits = %d; want an error and no request", err, hits.Load())
	}
}

func TestDevModeAllowList(t *testing.T) {
	srv, hits := countingServer(t, ok)
	dev := ssrf.Policy{Schemes: []string{"http"}, Allow: []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")}}
	resp, err := dev.Client(2 * time.Second).Get(srv.URL)
	if err != nil {
		t.Fatalf("dev policy: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK || hits.Load() != 1 {
		t.Fatalf("status %d, hits %d; want 200 and one request", resp.StatusCode, hits.Load())
	}
}

func TestRedirectsAreRechecked(t *testing.T) {
	dev := ssrf.Policy{Schemes: []string{"http"}, Allow: []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")}}
	for name, target := range map[string]string{
		"to the metadata address": "http://169.254.169.254/latest/meta-data/iam/",
		"to another scheme":       "ftp://files.example.com/",
		"to credentials":          "http://admin:admin@127.0.0.1/",
	} {
		t.Run(name, func(t *testing.T) {
			srv, _ := countingServer(t, func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, target, http.StatusFound)
			})
			_, err := dev.Client(2 * time.Second).Get(srv.URL)
			if !errors.Is(err, ssrf.ErrBlocked) {
				t.Fatalf("redirect %s: error = %v; want ErrBlocked", target, err)
			}
		})
	}
	t.Run("too many hops", func(t *testing.T) {
		var srv *httptest.Server
		srv, _ = countingServer(t, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, srv.URL+"/again", http.StatusFound)
		})
		_, err := dev.Client(2 * time.Second).Get(srv.URL)
		if err == nil || !strings.Contains(err.Error(), "redirects") {
			t.Fatalf("error = %v; want the redirect limit", err)
		}
	})
}

func TestClientIgnoresProxyEnvironment(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
	tr, isTransport := httpPolicy.Client(time.Second).Transport.(*http.Transport)
	if !isTransport || tr.Proxy != nil {
		t.Fatal("the transport must not use a proxy: the dialer would check the proxy, not the target")
	}
}
