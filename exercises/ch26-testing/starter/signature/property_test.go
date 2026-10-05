package signature_test

import (
	"testing"
)

// Property tests for the signer with pgregory.net/rapid (day 80, step 8).

// Round trip: whatever the secret, body and time, what Sign produces Verify accepts.
func TestPropertySignVerifyRoundTrip(t *testing.T) {
	t.Fatal("TODO: write the round-trip property; README day 80, step 8")
}

// Tampering: changing one byte of the body, or the secret, is always signature_mismatch.
func TestPropertyTamperingIsDetected(t *testing.T) {
	t.Fatal("TODO: write the tampering property; README day 80, step 8")
}
