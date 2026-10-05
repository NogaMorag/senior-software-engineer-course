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

	"example.com/ch26-testing/contract"
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

// The shared examples are the contract's own test cases. The Python consumer runs the same files.
func TestSchemaExamples(t *testing.T) {
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

// Producer side: what this package actually sends conforms to the schema.
func TestProducedBodiesConform(t *testing.T) {
	sch := compile(t)
	at := time.Date(2026, 10, 5, 12, 30, 0, 123456789, time.FixedZone("IDT", 3*60*60))
	tests := []struct {
		name string
		typ  string
		data any
	}{
		{"struct data", "invoice.paid", struct {
			AmountCents int `json:"amount_cents"`
		}{1999}},
		{"map data", "customer.address_changed", map[string]any{"city": "חיפה", "tags": []string{"vip"}}},
		{"empty object", "ping.sent", map[string]any{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev, err := contract.New("6f1c2a9e-3b7d-4c1a-9e2f-1a2b3c4d5e6f", tc.typ, at, tc.data)
			if err != nil {
				t.Fatal(err)
			}
			body, err := ev.Body()
			if err != nil {
				t.Fatal(err)
			}
			if err := validate(sch, body); err != nil {
				t.Fatalf("produced body %s does not conform: %v", body, err)
			}
		})
	}
}

func TestNewRefusesWhatTheSchemaRefuses(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		typ  string
		data any
	}{
		{"paid", map[string]int{}},
		{"Invoice.Paid", map[string]int{}},
		{"invoice.paid\n", map[string]int{}},
		{"invoice.paid", []int{1999}},
		{"invoice.paid", nil},
	} {
		if _, err := contract.New("6f1c2a9e-3b7d-4c1a-9e2f-1a2b3c4d5e6f", tc.typ, now, tc.data); err == nil {
			t.Errorf("New(%q, %v) accepted it", tc.typ, tc.data)
		}
	}
}

// uuidGen draws a random version-4 UUID in its canonical text form.
var uuidGen = rapid.Custom(func(t *rapid.T) string {
	b := rapid.SliceOfN(rapid.Byte(), 16, 16).Draw(t, "uuid bytes")
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
})

// Producer and contract agree on every type string: New accepts a type exactly when the schema
// accepts a body with that type. Half the draws match the pattern, half are arbitrary strings.
func TestPropertyNewAgreesWithSchema(t *testing.T) {
	sch := compile(t)
	typeGen := rapid.OneOf(rapid.StringMatching(`[a-z][a-z0-9_]{0,8}(\.[a-z][a-z0-9_]{0,8}){1,3}`), rapid.String())
	rapid.Check(t, func(t *rapid.T) {
		id := uuidGen.Draw(t, "id")
		typ := typeGen.Draw(t, "type")
		at := time.Unix(rapid.Int64Range(0, 4102444800).Draw(t, "unix"), 0) // 1970..2100
		data := rapid.MapOf(rapid.StringN(1, 10, 40), rapid.Int()).Draw(t, "data")

		ev, newErr := contract.New(id, typ, at, data)
		hand, err := json.Marshal(map[string]any{"id": id, "type": typ, "created_at": at.UTC().Format(time.RFC3339), "data": data})
		if err != nil {
			t.Fatal(err)
		}
		schemaErr := validate(sch, hand)
		if (newErr == nil) != (schemaErr == nil) {
			t.Fatalf("type %q: New says %v, schema says %v", typ, newErr, schemaErr)
		}
		if newErr != nil {
			return
		}
		body, err := ev.Body()
		if err != nil {
			t.Fatal(err)
		}
		if err := validate(sch, body); err != nil {
			t.Fatalf("produced body %s does not conform: %v", body, err)
		}
	})
}
