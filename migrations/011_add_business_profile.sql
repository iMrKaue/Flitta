BEGIN;

ALTER TABLE clients
    ADD COLUMN description TEXT,
    ADD COLUMN address TEXT,
    ADD COLUMN city VARCHAR(120),
    ADD COLUMN instagram VARCHAR(100),
    ADD COLUMN welcome_message TEXT;

COMMIT;
