package signature_test

import (
	"testing"
)

// FuzzParseHeader feeds ParseHeader arbitrary strings (day 81, step 1).
//
//	go test -run '^$' -fuzz '^FuzzParseHeader$' -fuzztime 30s ./signature
//
// TODO: add every header from testdata/signature-vectors.json as a seed with f.Add, then in
// f.Fuzz check what must hold for any input: no panic, and on success a timestamp >= 0, at
// least one 32-byte signature, and ParseHeader(h.String()) equal to h.
func FuzzParseHeader(f *testing.F) {
	f.Add("t=1791201600")
	f.Fuzz(func(t *testing.T, header string) {
		t.Fatal("TODO: write the fuzz target; README day 81, step 1")
	})
}
