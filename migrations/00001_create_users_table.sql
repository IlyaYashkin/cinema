-- +goose up
CREATE SCHEMA sso;

CREATE TABLE sso.roles
(
    id   SMALLINT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

INSERT INTO sso.roles (id, name)
VALUES (1, 'member'),
       (2, 'admin');

CREATE TABLE sso.users
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT UNIQUE NOT NULL,
    role_id       SMALLINT NOT NULL REFERENCES sso.roles(id) DEFAULT 1,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ      DEFAULT now()
);

-- +goose down
DROP TABLE sso.users;

DROP SCHEMA sso;
