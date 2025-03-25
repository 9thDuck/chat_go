CREATE TABLE IF NOT EXISTS encryption_keys (
    user_id BIGSERIAL REFERENCES users(id) ON DELETE CASCADE,
    key_id VARCHAR(100) UNIQUE NOT NULL,
    key VARCHAR(100) UNIQUE NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, key_id, key)
);