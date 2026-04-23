CREATE TABLE IF NOT EXISTS catalog (
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    usage_count BIGINT NOT NULL DEFAULT 0,
    last_used TIMESTAMPTZ,
    success_count BIGINT NOT NULL DEFAULT 0,
    error_count BIGINT NOT NULL DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (name, type)
);
