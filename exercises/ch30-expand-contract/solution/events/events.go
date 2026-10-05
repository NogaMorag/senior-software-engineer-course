// Package events is the same rename on the wire: the user.renamed event goes from a fullname
// field (version 1) to display_name (version 2) without breaking producers or consumers that
// are deployed at different times.
package events

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Envelope is what every event carries. Readers dispatch on Type and Version and never
// guess the shape of Data.
type Envelope struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Version int             `json:"version"`
	Data    json.RawMessage `json:"data"`
}

// RenamedV1 is the payload old producers send and old consumers expect.
type RenamedV1 struct {
	UserID   int64  `json:"user_id"`
	Fullname string `json:"fullname"`
}

// RenamedV2 is the new payload. During the transition producers also fill Fullname, so that
// consumers still on version 1 keep working: an additive change, the expand step on the wire.
type RenamedV2 struct {
	UserID      int64  `json:"user_id"`
	DisplayName string `json:"display_name"`
	Fullname    string `json:"fullname,omitempty"` // drop when no version 1 consumer is left
}

// ErrUnknownVersion means the event is newer than this reader. The caller decides: park it in a
// dead-letter queue and alert, never drop it silently.
var ErrUnknownVersion = errors.New("unknown event version")

// NewRenamed builds a version 2 event. While old consumers exist, dual is true and the event
// carries both fields.
func NewRenamed(id string, userID int64, name string, dual bool) (Envelope, error) {
	p := RenamedV2{UserID: userID, DisplayName: name}
	if dual {
		p.Fullname = name
	}
	data, err := json.Marshal(p)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{ID: id, Type: "user.renamed", Version: 2, Data: data}, nil
}

// NameV1 is the old consumer: it knows only version 1. encoding/json ignores fields it does not
// know, so it reads a dual-written version 2 payload too. That is the tolerant reader.
func NameV1(e Envelope) (string, error) {
	var p RenamedV1
	if err := json.Unmarshal(e.Data, &p); err != nil {
		return "", fmt.Errorf("user.renamed v%d: %w", e.Version, err)
	}
	if p.Fullname == "" {
		return "", fmt.Errorf("user.renamed v%d: no fullname", e.Version)
	}
	return p.Fullname, nil
}

// Name is the new consumer: it reads both versions, so events written before the change (and
// replayed from a topic months later) still work. That is backward compatibility in time.
func Name(e Envelope) (string, error) {
	switch e.Version {
	case 1:
		var p RenamedV1
		if err := json.Unmarshal(e.Data, &p); err != nil {
			return "", fmt.Errorf("user.renamed v1: %w", err)
		}
		return p.Fullname, nil
	case 2:
		var p RenamedV2
		if err := json.Unmarshal(e.Data, &p); err != nil {
			return "", fmt.Errorf("user.renamed v2: %w", err)
		}
		return p.DisplayName, nil
	}
	return "", fmt.Errorf("user.renamed v%d: %w", e.Version, ErrUnknownVersion)
}
