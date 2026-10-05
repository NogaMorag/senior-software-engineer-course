package platform_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/hookline/internal/platform"
)

func TestDebugHandlerServesPprof(t *testing.T) {
	srv := httptest.NewServer(platform.DebugHandler())
	defer srv.Close()

	for _, path := range []string{"/debug/pprof/", "/debug/pprof/goroutine?debug=1", "/debug/pprof/cmdline"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || len(body) == 0 {
			t.Errorf("GET %s = %d with %d bytes", path, resp.StatusCode, len(body))
		}
		if path == "/debug/pprof/" && !strings.Contains(string(body), "goroutine") {
			t.Errorf("index page does not list the goroutine profile")
		}
	}
}
