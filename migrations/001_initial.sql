CREATE TABLE IF NOT EXISTS test_runs (
    id BIGSERIAL PRIMARY KEY,
    test_id TEXT NOT NULL,
    name TEXT NOT NULL,
    commit_sha TEXT,
    branch TEXT,
    runner TEXT,
    status TEXT NOT NULL CHECK (status IN ('passed', 'failed', 'skipped', 'error')),
    attempt INTEGER NOT NULL DEFAULT 1,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    ts TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_test_runs_test_id_ts ON test_runs (test_id, ts DESC);
CREATE INDEX IF NOT EXISTS idx_test_runs_commit_test ON test_runs (commit_sha, test_id);
CREATE INDEX IF NOT EXISTS idx_test_runs_status ON test_runs (status);
