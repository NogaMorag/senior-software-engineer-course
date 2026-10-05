package redact_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"example.com/ch28-security/internal/redact"
)

// Placeholder values: each must be absent from every log line.
const (
	password  = "change-me-password"
	bearerTok = "eyJhbGciOiJSUzI1NiJ9.change-me.sig"
	apiKey    = "hl_4f9c2a7d1e03_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	signing   = "change-me-signing-secret"
	cookie    = "session=change-me-cookie"
)

func newLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{ReplaceAttr: redact.ReplaceAttr}))
}

func TestReplaceAttrRedactsSecrets(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(&buf)

	h := http.Header{}
	h.Set("Authorization", "Bearer "+bearerTok)
	h.Set("Cookie", cookie)
	h.Set("Content-Type", "application/json")

	log.Info("by key", slog.String("password", password), slog.String("Authorization", "Bearer "+bearerTok))
	log.Info("key spelling", slog.String("X-Api-Key", apiKey), slog.String("db_password", password))
	log.Info("in a group", slog.Group("endpoint", slog.String("url", "https://hooks.example.com"), slog.String("signing_secret", signing)))
	log.Info("by shape", slog.String("note", "retrying with Bearer "+bearerTok), slog.String("input", "key="+apiKey))
	log.Info("Secret type", slog.Any("cfg", redact.Secret(signing)))
	log.Info("headers", redact.Header("headers", h))
	log.With(slog.String("token", bearerTok)).Info("via With")

	out := buf.String()
	for _, secret := range []string{password, bearerTok, apiKey, signing, cookie} {
		if strings.Contains(out, secret) {
			t.Errorf("log output contains %q:\n%s", secret, out)
		}
	}
	// What is not secret must survive, or people turn the redactor off.
	for _, keep := range []string{`"url":"https://hooks.example.com"`, `"Content-Type":"application/json"`, `"note":"retrying with [REDACTED]"`} {
		if !strings.Contains(out, keep) {
			t.Errorf("log output lost %s:\n%s", keep, out)
		}
	}
	for i, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if !json.Valid([]byte(line)) {
			t.Errorf("line %d is not JSON: %s", i+1, line)
		}
	}
}

func TestSecretNeverPrints(t *testing.T) {
	s := redact.Secret(signing)
	type config struct {
		Name   string
		Secret redact.Secret
	}
	c := config{Name: "hookline", Secret: s}
	js, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, form := range []string{fmt.Sprint(s), fmt.Sprintf("%v %s %+v %#v", s, s, c, c), string(js)} {
		if strings.Contains(form, signing) {
			t.Errorf("printed form leaks the secret: %s", form)
		}
	}
	if s.Reveal() != signing {
		t.Error("Reveal must return the value")
	}
}

// TestRawMapIsNotRedacted documents the limit: ReplaceAttr does not look
// inside a map, so an http.Header logged with slog.Any leaks. Use redact.Header.
func TestRawMapIsNotRedacted(t *testing.T) {
	var buf bytes.Buffer
	h := http.Header{"Authorization": {"Bearer " + bearerTok}}
	newLogger(&buf).Info("raw", slog.Any("headers", h))
	if !strings.Contains(buf.String(), bearerTok) {
		t.Fatal("expected the documented leak; if slog now descends into maps, update the package doc")
	}
}
