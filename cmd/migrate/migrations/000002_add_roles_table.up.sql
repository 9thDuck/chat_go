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
    'A user can create message and group chats',
    1
);
INSERT INTO
    roles (name, description, level)
VALUES(
    'moderator',
    'A moderator can delete messages in group chats',
    2
);
INSERT INTO
    roles (name, description, level)
VALUES(
    'admin',
    'admin can delete users',
    3
);