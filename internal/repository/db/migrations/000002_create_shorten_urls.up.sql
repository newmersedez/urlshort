CREATE TABLE shorten_urls (
    id              VARCHAR(256) PRIMARY KEY,
    user_id         VARCHAR(256) NOT NULL,
    original_value  VARCHAR(2048) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ix_shorten_urls_original_value ON shorten_urls(original_value);
CREATE INDEX IF NOT EXISTS ix_shorten_urls_user_id ON shorten_urls(user_id);