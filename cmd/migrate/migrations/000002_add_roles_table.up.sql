CREATE TABLE IF NOT EXISTS roles(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level int NOT NULL DEFAULT 0,
    description TEXT
);
-- Default roles insertion
INSERT INTO
    roles (name, description, level)
VALUES(
    'user',
    'A user can create posts and comments',
    1
);
INSERT INTO
    roles (name, description, level)
VALUES(
    'moderator',
    'A moderator can update other user posts',
    2
);
INSERT INTO
    roles (name, description, level)
VALUES(
    'admin',
    'nd admin can update and delete other user posts',
    3
);