package events_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"example.com/ch30-expand-contract/events"
)

func TestConsumersAcrossVersions(t *testing.T) {
	v1 := events.Envelope{ID: "e1", Type: "user.renamed", Version: 1,
		Data: json.RawMessage(`{"user_id":7,"fullname":"Ada Lovelace"}`)}
	dual, err := events.NewRenamed("e2", 7, "Ada Lovelace", true)
	if err != nil {
		t.Fatal(err)
	}
	newOnly, err := events.NewRenamed("e3", 7, "Ada Lovelace", false)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		read    func(events.Envelope) (string, error)
		event   events.Envelope
		wantErr bool
	}{
		{"old consumer, old event", events.NameV1, v1, false},
		{"old consumer, dual-written v2 (tolerant reader)", events.NameV1, dual, false},
		{"old consumer, v2 without fullname: contract too early", events.NameV1, newOnly, true},
		{"new consumer, old event from the topic", events.Name, v1, false},
		{"new consumer, dual-written v2", events.Name, dual, false},
		{"new consumer, v2 after the contract", events.Name, newOnly, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.read(tt.event)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("got %q, want an error", got)
				}
				return
			}
			if err != nil || got != "Ada Lovelace" {
				t.Fatalf("got %q, %v; want %q", got, err, "Ada Lovelace")
			}
		})
	}
}

func TestUnknownVersionIsAnErrorNotAGuess(t *testing.T) {
	e := events.Envelope{ID: "e9", Type: "user.renamed", Version: 3, Data: json.RawMessage(`{}`)}
	if _, err := events.Name(e); !errors.Is(err, events.ErrUnknownVersion) {
		t.Fatalf("version 3: err = %v, want ErrUnknownVersion", err)
	}
}

// TestStrictReaderBreaksOnAdditiveChange is why consumers must be tolerant: a decoder that
// rejects unknown fields fails on the dual-written payload, which only added a field.
func TestStrictReaderBreaksOnAdditiveChange(t *testing.T) {
	dual, err := events.NewRenamed("e2", 7, "Ada Lovelace", true)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(dual.Data))
	dec.DisallowUnknownFields()
	var p events.RenamedV1
	if err := dec.Decode(&p); err == nil {
		t.Fatal("a strict version 1 reader accepted a payload with display_name")
	} else {
		t.Logf("strict reader: %v", err)
	}
}
