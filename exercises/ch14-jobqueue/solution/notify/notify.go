// Package notify is an idempotent job handler: it sends a message through an external
// provider and records it, so that running the same job twice sends once.
package notify

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch14-jobqueue/queue"
)

// Schema creates the notifications table.
//
//go:embed schema.sql
var Schema string

// Kind is the job kind this handler serves.
const Kind = "notify"

// Payload is the job's input.
type Payload struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

// Sender is the external provider. Like most real providers, it accepts an idempotency
// key and promises that two calls with the same key deliver one message.
type Sender interface {
	Send(ctx context.Context, idempotencyKey, user, message string) error
}

// Handler builds the job handler.
//
// Two defences, because a crash can land between any two lines:
//  1. Before sending, look for our own record of this job. Found: it is a re-run of a
//     job that already finished its work, so skip.
//  2. The provider call carries a key derived from the job id. A crash after Send and
//     before the insert means the re-run sends again, with the same key, and the
//     provider drops the duplicate.
func Handler(db *pgxpool.Pool, s Sender) func(context.Context, queue.Job) error {
	return func(ctx context.Context, j queue.Job) error {
		var p Payload
		if err := json.Unmarshal(j.Payload, &p); err != nil || p.User == "" {
			return queue.Permanent(fmt.Errorf("bad notify payload %s", j.Payload))
		}
		key := fmt.Sprintf("notify-job-%d", j.ID) // stable across attempts and replays

		var done bool
		if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM notifications WHERE idempotency_key = $1)`, key).Scan(&done); err != nil {
			return err
		}
		if done {
			return nil
		}
		if err := s.Send(ctx, key, p.User, p.Message); err != nil {
			return fmt.Errorf("send to %s: %w", p.User, err)
		}
		_, err := db.Exec(ctx, `
			INSERT INTO notifications (idempotency_key, user_name, message) VALUES ($1, $2, $3)
			ON CONFLICT (idempotency_key) DO NOTHING`, key, p.User, p.Message)
		return err
	}
}
