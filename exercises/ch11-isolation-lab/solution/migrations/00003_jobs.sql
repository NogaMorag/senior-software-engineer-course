-- +goose Up
CREATE TABLE jobs (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    payload    text NOT NULL,
    status     text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'done')),
    done_by    text,
    done_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- Workers only ever look for queued jobs, oldest first.
CREATE INDEX jobs_queued_idx ON jobs (id) WHERE status = 'queued';

-- +goose Down
DROP TABLE jobs;
