ALTER TABLE IF EXISTS users
ADD COLUMN role_id INT NOT NULL REFERENCES roles(id);

UPDATE users
SET role_id= (
    SELECT id FROM roles WHERE name='user'
);
