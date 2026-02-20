CREATE TABLE shorten_urls (
    id              VARCHAR(256) PRIMARY KEY,
    original_value  VARCHAR(2048) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ix_shorten_urls_original_value ON shorten_urls(original_value);