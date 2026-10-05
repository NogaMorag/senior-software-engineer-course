// Package delivery sends deliveries (milestone M5): the signer, the HTTP sender and its retry
// classification, the queue operations on the deliveries table, and the worker pool.
package delivery

import ()

// Header names every delivery carries (spec F4).
const (
	HeaderEventID    = "Hookline-Event-Id"
	HeaderDeliveryID = "Hookline-Delivery-Id"
	HeaderAttempt    = "Hookline-Attempt"
	HeaderSignature  = "Hookline-Signature"
)

// Sign returns the Hookline-Signature value for body sent at unix second ts (spec F5):
// t=<ts>,v1=<lowercase hex HMAC-SHA256(secret, "<ts>." + body)>. The receiver recomputes the HMAC over
// the exact raw bytes, so the body must be sent exactly as signed. The shared test vectors
// (sink/data/signature-vectors.json) pin the format for Go and Python alike.
func Sign(secret string, ts int64, body []byte) string {
	panic("TODO: see chapter 38")

}
