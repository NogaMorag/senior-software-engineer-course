-- +goose Up
-- Milestone M8: the W3C traceparent of the ingest request, so that the relay, the dispatcher and the
-- workers, which handle the event later and in other processes, continue the same trace (chapter 24).
-- NULL for events ingested without telemetry.
ALTER TABLE events ADD COLUMN trace_context text;

-- +goose Down
ALTER TABLE events DROP COLUMN trace_context;
