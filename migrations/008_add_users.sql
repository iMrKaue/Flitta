BEGIN;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,

    client_id INTEGER NOT NULL,

    name VARCHAR(150) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,

    role VARCHAR(30) NOT NULL DEFAULT 'owner',
    active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP WITHOUT TIME ZONE
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP WITHOUT TIME ZONE
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT users_client_id_fkey
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE,

    CONSTRAINT users_name_not_blank
        CHECK (BTRIM(name) <> ''),

    CONSTRAINT users_email_not_blank
        CHECK (BTRIM(email) <> ''),

    CONSTRAINT users_role_check
        CHECK (role IN ('owner', 'admin', 'reception')) 
);

CREATE UNIQUE INDEX users_email_normalized_unique
    ON users (LOWER(BTRIM(email)));

INSERT INTO users (
    client_id,
    name,
    email,
    password,
    role,
    active
)
SELECT
    id,
    name,
    LOWER(BTRIM(email)),
    password,
    'owner',
    TRUE
FROM clients;

COMMIT;
