CREATE INDEX idx_messages_sender_id ON messages (sender_id);
CREATE INDEX idx_messages_receiver_id ON messages (receiver_id);
CREATE INDEX idx_messages_is_delivered ON messages (is_delivered);
CREATE INDEX idx_messages_created_at ON messages (created_at);