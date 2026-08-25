BEGIN;

ALTER TABLE services
    ALTER COLUMN client_id SET NOT NULL;

ALTER TABLE appointments
    ALTER COLUMN client_id SET NOT NULL;

ALTER TABLE working_hours
    ALTER COLUMN client_id SET NOT NULL;

ALTER TABLE services
    ADD CONSTRAINT services_client_id_fkey
    FOREIGN KEY (client_id)
    REFERENCES clients(id)
    ON DELETE RESTRICT;

ALTER TABLE appointments
    ADD CONSTRAINT appointments_client_id_fkey
    FOREIGN KEY (client_id)
    REFERENCES clients(id)
    ON DELETE RESTRICT;

ALTER TABLE working_hours
    ADD CONSTRAINT working_hours_client_id_fkey
    FOREIGN KEY (client_id)
    REFERENCES clients(id)
    ON DELETE RESTRICT;

COMMIT;
