CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE shorten_urls
ADD user_id UUID CONSTRAINT ix_shorten_urls_user_id DEFAULT uuid_generate_v4();