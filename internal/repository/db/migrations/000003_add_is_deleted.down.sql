DROP INDEX IF EXISTS ix_shorten_urls_is_deleted;

ALTER TABLE shorten_urls DROP COLUMN is_deleted;
