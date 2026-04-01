-- 001_initial_schema
-- Core tables for the Jarvis personal assistant (PostgreSQL).

-------------------------------------------------------
-- MEMORIES
-------------------------------------------------------

CREATE TABLE IF NOT EXISTS memories (
    id         SERIAL PRIMARY KEY,
    content    TEXT   NOT NULL,
    tags       JSONB  DEFAULT '[]'::jsonb,
    embedding  JSONB  DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_memories_created
    ON memories(created_at DESC);

-- GIN index on tags for fast JSONB containment queries
-- e.g. WHERE tags @> '["finance"]'
CREATE INDEX IF NOT EXISTS idx_memories_tags
    ON memories USING GIN (tags);

-------------------------------------------------------
-- FULL-TEXT SEARCH (spanish dictionary)
-- Uses tsvector/tsquery instead of FTS5.
-------------------------------------------------------

-- Stored generated column for the search vector
ALTER TABLE memories
    ADD COLUMN IF NOT EXISTS search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('spanish', content)) STORED;

CREATE INDEX IF NOT EXISTS idx_memories_fts
    ON memories USING GIN (search_vector);

-------------------------------------------------------
-- CONVERSATIONS
-------------------------------------------------------

CREATE TABLE IF NOT EXISTS conversations (
    id         SERIAL PRIMARY KEY,
    session_id TEXT NOT NULL,
    role       TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_session
    ON conversations(session_id, created_at);

-- Partial index for fast "latest messages" queries per session
CREATE INDEX IF NOT EXISTS idx_conversations_session_recent
    ON conversations(session_id, created_at DESC);
