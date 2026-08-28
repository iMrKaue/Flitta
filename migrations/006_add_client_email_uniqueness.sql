BEGIN;

UPDATE clients
SET email = LOWER(BTRIM(email))
WHERE email IS NOT NULL;

ALTER TABLE clients
    ALTER COLUMN email SET NOT NULL;

ALTER TABLE clients
    ADD CONSTRAINT clients_email_not_blank
    CHECK (BTRIM(email) <> '');

CREATE UNIQUE INDEX clients_email_normalized_unique
    ON clients (LOWER(BTRIM(email)));

COMMIT;