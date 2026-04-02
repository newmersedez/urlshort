ALTER TABLE shorten_urls ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS ix_shorten_urls_is_deleted ON shorten_urls(is_deleted);
