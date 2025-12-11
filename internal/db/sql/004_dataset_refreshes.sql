CREATE TABLE IF NOT EXISTS dataset_refreshes (
    dataset TEXT PRIMARY KEY,
    last_loaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    row_count BIGINT NOT NULL DEFAULT 0,
    notes TEXT
);
