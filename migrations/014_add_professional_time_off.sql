BEGIN;

CREATE TABLE professional_time_off (
    id SERIAL PRIMARY KEY,

    client_id INTEGER NOT NULL,

    professional_id INTEGER NOT NULL,

    start_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    end_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    reason VARCHAR(255),

    created_at TIMESTAMP WITHOUT TIME ZONE
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT professional_time_off_range_check
        CHECK (
            start_at < end_at
        ),

    CONSTRAINT professional_time_off_client_fkey
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE,

    CONSTRAINT professional_time_off_professional_fkey
        FOREIGN KEY (
            professional_id,
            client_id
        )
        REFERENCES professionals(
            id,
            client_id
        )
        ON DELETE CASCADE
);

CREATE INDEX idx_professional_time_off_lookup
    ON professional_time_off (
        client_id,
        professional_id,
        start_at,
        end_at
    );

COMMIT;

