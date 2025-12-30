CREATE TABLE IF NOT EXISTS comments (
    cid SERIAL PRIMARY KEY,
    pid INT NULL,
    content TEXT NOT NULL,
    content_tsv tsvector GENERATED ALWAYS AS (
        to_tsvector('simple', content)
    ) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    author TEXT,
    CONSTRAINT fk_comments_parent FOREIGN KEY (pid) REFERENCES comments (cid) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_comments_content_fts ON comments USING GIN (content_tsv);

CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments (deleted_at);

CREATE INDEX IF NOT EXISTS idx_comments_pid ON comments (pid);

CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments (created_at);