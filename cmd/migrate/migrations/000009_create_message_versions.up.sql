CREATE TABLE IF NOT EXISTS message_versions (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGSERIAL NOT NULL REFERENCES messages(id),
    version INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);