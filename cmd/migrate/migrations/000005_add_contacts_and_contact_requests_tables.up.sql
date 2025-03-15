CREATE TABLE IF NOT EXISTS contacts (
    user_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, contact_id)
);

CREATE INDEX ON contacts (user_id, contact_id);

CREATE TABLE IF NOT EXISTS contact_requests (
    sender_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (sender_id, receiver_id)
);

CREATE INDEX ON contact_requests (sender_id, receiver_id);