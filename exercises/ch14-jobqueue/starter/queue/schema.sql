-- Chapter 14 schema: a job queue on PostgreSQL.

CREATE TABLE jobs (
    id            bigserial PRIMARY KEY,
    kind          text        NOT NULL,                 -- which handler runs it
    payload       jsonb       NOT NULL DEFAULT '{}',
    unique_key    text        UNIQUE,                   -- optional: enqueue at most once per key
    state         text        NOT NULL DEFAULT 'pending'
                  CHECK (state IN ('pending', 'running', 'done', 'dead')),
    run_at        timestamptz NOT NULL DEFAULT now(),   -- not before this time
    attempts      int         NOT NULL DEFAULT 0,       -- claims so far; also the fencing token
    max_attempts  int         NOT NULL DEFAULT 5 CHECK (max_attempts > 0),
    locked_by     text,                                 -- worker holding the lease
    locked_at     timestamptz,
    locked_until  timestamptz,                          -- lease end; after it, any worker may reclaim
    last_error    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    finished_at   timestamptz
);

-- Workers look for due pending jobs, and for running jobs whose lease ran out.
CREATE INDEX jobs_due    ON jobs (run_at, id)    WHERE state = 'pending';
CREATE INDEX jobs_leased ON jobs (locked_until)  WHERE state = 'running';

-- Append-only history: one row per finished attempt, kept across replays.
CREATE TABLE job_attempts (
    id           bigserial PRIMARY KEY,
    job_id       bigint      NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    attempt      int         NOT NULL,
    worker       text,
    started_at   timestamptz NOT NULL,
    finished_at  timestamptz NOT NULL DEFAULT now(),
    error        text                                   -- NULL means success
);
CREATE INDEX job_attempts_job ON job_attempts (job_id, id);
