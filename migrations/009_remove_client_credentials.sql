BEGIN;

DROP INDEX clients_email_normalized_unique;

ALTER TABLE clients
    DROP CONSTRAINT clients_email_not_blank;

ALTER TABLE clients
    DROP COLUMN email,
    DROP COLUMN password;

COMMIT;
