package signature_test

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"example.com/ch26-testing/signature"
)

// vectors mirrors testdata/signature-vectors.json, the file the Python receiver also passes.
type vectors struct {
	ToleranceSeconds int `json:"tolerance_seconds"`
	Sign             []struct {
		Name      string `json:"name"`
		Secret    string `json:"secret"`
		Timestamp int64  `json:"timestamp"`
		Body      string `json:"body"`
		Header    string `json:"header"`
	} `json:"sign"`
	Verify []struct {
		Name   string `json:"name"`
		Secret string `json:"secret"`
		Header string `json:"header"`
		Body   string `json:"body"`
		Now    int64  `json:"now"`
		Valid  bool   `json:"valid"`
		Reason string `json:"reason"`
	} `json:"verify"`
}

func loadVectors(t *testing.T) vectors {
	t.Helper()
	raw, err := os.ReadFile("testdata/signature-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var v vectors
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestSignVectors(t *testing.T) {
	for _, tc := range loadVectors(t).Sign {
		t.Run(tc.Name, func(t *testing.T) {
			if got := signature.Sign([]byte(tc.Secret), tc.Timestamp, []byte(tc.Body)); got != tc.Header {
				t.Fatalf("Sign = %q, want %q", got, tc.Header)
			}
		})
	}
}

func TestVerifyVectors(t *testing.T) {
	v := loadVectors(t)
	tolerance := time.Duration(v.ToleranceSeconds) * time.Second
	for _, tc := range v.Verify {
		t.Run(tc.Name, func(t *testing.T) {
			err := signature.Verify(tc.Header, []byte(tc.Body), []byte(tc.Secret), time.Unix(tc.Now, 0), tolerance)
			if tc.Valid {
				if err != nil {
					t.Fatalf("Verify = %v, want valid", err)
				}
				return
			}
			var serr *signature.Error
			if !errors.As(err, &serr) {
				t.Fatalf("Verify = %v, want a *signature.Error with reason %s", err, tc.Reason)
			}
			if string(serr.Reason) != tc.Reason {
				t.Fatalf("reason = %s (%v), want %s", serr.Reason, err, tc.Reason)
			}
		})
	}
}
