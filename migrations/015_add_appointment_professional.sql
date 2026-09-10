BEGIN;

ALTER TABLE appointments
    ADD COLUMN professional_id INTEGER;

ALTER TABLE appointments
    ADD CONSTRAINT appointments_professional_fkey
    FOREIGN KEY (
        professional_id,
        client_id
    )
    REFERENCES professionals(
        id,
        client_id
    )
    ON DELETE RESTRICT;

DROP INDEX IF EXISTS unique_schedule_active;

CREATE UNIQUE INDEX unique_schedule_active_professional
    ON appointments (
        client_id,
        professional_id,
        date,
        time
    )
    WHERE status <> 'cancelled'
      AND professional_id IS NOT NULL;

CREATE UNIQUE INDEX unique_schedule_active_legacy
    ON appointments (
        client_id,
        date,
        time
    )
    WHERE status <> 'cancelled'
      AND professional_id IS NULL;

COMMIT;
