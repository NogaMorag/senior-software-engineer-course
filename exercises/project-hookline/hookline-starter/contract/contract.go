// Package contract holds the webhook contract (milestone M9): webhook-event.schema.json is the body
// every receiver gets, and testdata/examples are its shared test cases. The producer is
// delivery.Job.Body; the Go tests check that it conforms, and the sink's Python tests check the same
// examples from the receiver's side. Neither side imports the other: they share only these files.
package contract

import _ "embed"

// Schema is the JSON Schema (draft 2020-12) of a webhook body.
//
//go:embed webhook-event.schema.json
var Schema []byte

// SchemaURL is the schema's $id; validators use it as the schema's name.
const SchemaURL = "https://example.com/hookline/webhook-event.schema.json"
