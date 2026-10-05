package delivery_test

import (
	"encoding/json"
	"os"
	"testing"

	"example.com/hookline/internal/delivery"
)

// TestSignMatchesSharedVectors: the Go signer and the Python verifier (sink/tests) read the same file,
// so a change to the format on one side fails a test on the other (spec F5). The verify vectors
// (tolerance, malformed headers) are the receiver's concern and are tested in Python.
func TestSignMatchesSharedVectors(t *testing.T) {
	raw, err := os.ReadFile("../../sink/data/signature-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Sign []struct {
			Name, Secret, Body, Header string
			Timestamp                  int64
		} `json:"sign"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Sign) == 0 {
		t.Fatal("no sign vectors")
	}
	for _, v := range file.Sign {
		t.Run(v.Name, func(t *testing.T) {
			if got := delivery.Sign(v.Secret, v.Timestamp, []byte(v.Body)); got != v.Header {
				t.Fatalf("Sign = %s\nwant   %s", got, v.Header)
			}
		})
	}
}
