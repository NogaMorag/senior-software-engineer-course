package contract_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"pgregory.net/rapid"

	"example.com/hookline/contract"
	"example.com/hookline/internal/delivery"
	"example.com/hookline/internal/domain"
)

// compile loads the embedded schema. AssertFormat matters: in draft 2020-12 "format" is only an
// annotation unless the validator is told to assert it, and then "evt_123" would pass as a uuid.
func compile(t testing.TB) *jsonschema.Schema {
	t.Helper()
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(contract.Schema))
	if err != nil {
		t.Fatalf("schema is not JSON: %v", err)
	}
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	if err := c.AddResource(contract.SchemaURL, doc); err != nil {
		t.Fatal(err)
	}
	sch, err := c.Compile(contract.SchemaURL)
	if err != nil {
		t.Fatalf("schema does not compile: %v", err)
	}
	return sch
}

func validate(sch *jsonschema.Schema, body []byte) error {
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("not JSON: %w", err)
	}
	return sch.Validate(inst)
}

// The shared examples are the contract's own test cases. The sink's Python tests run the same files.
func TestContractExamples(t *testing.T) {
	sch := compile(t)
	for _, kind := range []string{"valid", "invalid"} {
		files, err := filepath.Glob(filepath.Join("testdata", "examples", kind, "*.json"))
		if err != nil || len(files) == 0 {
			t.Fatalf("no %s examples (%v)", kind, err)
		}
		for _, file := range files {
			t.Run(kind+"/"+filepath.Base(file), func(t *testing.T) {
				body, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				err = validate(sch, body)
				if kind == "valid" && err != nil {
					t.Fatalf("rejected a valid example: %v", err)
				}
				if kind == "invalid" && err == nil {
					t.Fatal("accepted an invalid example")
				}
			})
		}
	}
}

// Producer side: the bytes the worker signs and sends conform to the schema.
func TestContractDeliveryBodyConforms(t *testing.T) {
	sch := compile(t)
	at := time.Date(2026, 10, 5, 12, 30, 0, 123456000, time.UTC)
	for _, payload := range []string{`{}`, `{"amount_cents":1999}`, `{"customer":{"city":"חיפה","tags":["vip"]}}`} {
		job := delivery.Job{EventID: domain.NewID(at), EventType: "invoice.paid", EventAt: at, Payload: json.RawMessage(payload)}
		if err := validate(sch, job.Body()); err != nil {
			t.Errorf("body %s does not conform: %v", job.Body(), err)
		}
	}
}

// The API and the contract agree on every type string: ingest accepts a type exactly when the schema
// accepts a body with that type, and every accepted one produces a conforming body. Half the draws match
// the pattern, half are arbitrary strings.
func TestContractPropertyIngestRuleAgreesWithSchema(t *testing.T) {
	sch := compile(t)
	typeGen := rapid.OneOf(rapid.StringMatching(`[a-z][a-z0-9_]{0,8}(\.[a-z][a-z0-9_]{0,8}){0,3}`), rapid.String())
	rapid.Check(t, func(t *rapid.T) {
		typ := typeGen.Draw(t, "type")
		at := time.Unix(rapid.Int64Range(0, 4102444800).Draw(t, "unix"), rapid.Int64Range(0, 999999).Draw(t, "micros")*1000).UTC()
		payload, err := json.Marshal(rapid.MapOf(rapid.StringN(1, 10, 40), rapid.Int()).Draw(t, "payload"))
		if err != nil {
			t.Fatal(err)
		}
		job := delivery.Job{EventID: domain.NewID(at), EventType: typ, EventAt: at, Payload: payload}

		ingestOK := domain.ValidateEventType(typ, false) == ""
		schemaErr := validate(sch, job.Body())
		if ingestOK != (schemaErr == nil) {
			t.Fatalf("type %q: ingest accepts it %v, schema says %v", typ, ingestOK, schemaErr)
		}
	})
}
