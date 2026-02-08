CREATE TABLE shorten_urls (
    id              TEXT PRIMARY KEY,
    original_value  TEXT NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);