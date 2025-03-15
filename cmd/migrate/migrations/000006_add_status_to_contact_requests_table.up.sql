ALTER TABLE contact_requests
ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'pending';

-- accepted, rejected, pending
