CREATE TABLE IF NOT EXISTS attachments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGSERIAL NOT NULL REFERENCES messages(id),
    path TEXT NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attachments_message_id ON attachments (message_id);