package api_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/hookline/internal/api"
	"example.com/hookline/internal/platform"
)

// lockedBuffer is a bytes.Buffer safe for the concurrent writes of handlers.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

// TestLogsNeverContainKeysOrSecrets is spec N7's log scan: every API path that handles an API key or a
// signing secret runs with the service's real logger at debug level, and neither value appears in any
// line, including one a careless change might add.
func TestLogsNeverContainKeysOrSecrets(t *testing.T) {
	var out lockedBuffer
	log := platform.NewLogger(&out, slog.LevelDebug)
	st := newStore(t)
	refuse := func(_ context.Context, raw string) error {
		if strings.Contains(raw, "internal") {
			return errors.New("ssrf: destination not allowed: " + raw)
		}
		return nil
	}
	hs := &harness{t: t, store: st, h: api.New(st, api.Options{Logger: log, CheckEndpoint: refuse, Now: (&clock{t: time.Now()}).now})}
	key := hs.tenant("Acme")

	created := hs.createEndpoint(key, "https://hooks.example.com/x")
	secret := created["secret"].(string)
	hs.do(req{Method: "GET", Path: "/v1/endpoints/" + created["id"].(string), Key: key})
	hs.do(req{Method: "POST", Path: "/v1/endpoints", Key: key, Body: `{"url":"https://internal.example/x"}`})
	hs.ingest(key, "scan-1", paid)
	wrong := "hk_" + strings.Repeat("Z", 43)
	hs.do(req{Method: "GET", Path: "/v1/tenant", Key: wrong})
	hs.do(req{Method: "GET", Path: "/v1/tenant", Header: map[string]string{"Authorization": "Bearer " + key + " trailing"}})
	// The careless line: logging the credentials outright. Redaction must still hold.
	log.Info("debugging auth with key "+key, slog.String("key", key), slog.String("secret", secret), slog.Any("endpoint", created))

	logs := out.String()
	if !strings.Contains(logs, `"msg":"request"`) {
		t.Fatalf("no access log lines captured:\n%s", logs)
	}
	for name, value := range map[string]string{"API key": key, "signing secret": secret, "rejected key": wrong} {
		if strings.Contains(logs, value) {
			t.Errorf("the logs contain the %s:\n%s", name, logs)
		}
	}
}
