BEGIN;

ALTER TABLE clients
    ADD COLUMN status VARCHAR(20)
        NOT NULL
        DEFAULT 'active',

    ADD COLUMN pilot_expires_at TIMESTAMPTZ;

ALTER TABLE clients
    ADD CONSTRAINT clients_status_check
    CHECK (status IN ('active', 'suspended'));

COMMIT;
