CREATE TABLE IF NOT EXISTS jobs (
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    queue        text        NOT NULL,
    payload      jsonb       NOT NULL,
    status       text        NOT NULL DEFAULT 'queued'
                 CHECK (status IN ('queued', 'running', 'done', 'failed')),
    attempts     int         NOT NULL DEFAULT 0,
    max_attempts int         NOT NULL DEFAULT 5,
    run_at       timestamptz NOT NULL DEFAULT now(),
    locked_until timestamptz,
    last_error   text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- Workers look for runnable jobs in one queue, oldest first. A partial
-- index keeps finished jobs out of it.
CREATE INDEX IF NOT EXISTS jobs_runnable ON jobs (queue, run_at, id)
    WHERE status IN ('queued', 'running');
