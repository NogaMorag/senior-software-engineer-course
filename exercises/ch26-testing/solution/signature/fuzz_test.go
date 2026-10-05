package signature_test

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"example.com/ch26-testing/signature"
)

// FuzzParseHeader feeds ParseHeader arbitrary strings. It checks what must hold for every input:
// no panic; on success, a sane result that formats and parses back to itself. The seed corpus is
// every header in the shared test vectors, plus a few shapes the vectors do not have.
//
//	go test -run '^$' -fuzz '^FuzzParseHeader$' -fuzztime 30s ./signature
//
// A failing input is saved under testdata/fuzz/FuzzParseHeader/; plain `go test` replays it from
// then on, so a crasher becomes a regression test once you commit the file.
func FuzzParseHeader(f *testing.F) {
	raw, err := os.ReadFile("testdata/signature-vectors.json")
	if err != nil {
		f.Fatal(err)
	}
	var v struct {
		Sign   []struct{ Header string } `json:"sign"`
		Verify []struct{ Header string } `json:"verify"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		f.Fatal(err)
	}
	for _, s := range v.Sign {
		f.Add(s.Header)
	}
	for _, s := range v.Verify {
		f.Add(s.Header)
	}
	f.Add("t=0,v1=") // a key with an empty value

	f.Fuzz(func(t *testing.T, header string) {
		h, err := signature.ParseHeader(header)
		if err != nil {
			return // rejecting is fine; panicking is not
		}
		if h.Timestamp < 0 || len(h.Signatures) == 0 {
			t.Fatalf("ParseHeader(%q) accepted %+v", header, h)
		}
		for _, s := range h.Signatures {
			if len(s) != sha256.Size {
				t.Fatalf("ParseHeader(%q): signature of %d bytes", header, len(s))
			}
		}
		again, err := signature.ParseHeader(h.String())
		if err != nil || !reflect.DeepEqual(again, h) {
			t.Fatalf("round trip of %q via %q: %+v, %v", header, h.String(), again, err)
		}
	})
}
