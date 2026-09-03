BEGIN;

CREATE TABLE professional_working_hours (
    id SERIAL PRIMARY KEY,

    client_id INTEGER NOT NULL,

    professional_id INTEGER NOT NULL,

    -- ISO weekday:
    -- 1 = segunda-feira
    -- 2 = terça-feira
    -- 3 = quarta-feira
    -- 4 = quinta-feira
    -- 5 = sexta-feira
    -- 6 = sábado
    -- 7 = domingo
    weekday SMALLINT NOT NULL,

    start_time TIME WITHOUT TIME ZONE NOT NULL,

    end_time TIME WITHOUT TIME ZONE NOT NULL,

    created_at TIMESTAMP WITHOUT TIME ZONE
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT professional_working_hours_weekday_check
        CHECK (
            weekday BETWEEN 1 AND 7
        ),

    CONSTRAINT professional_working_hours_time_check
        CHECK (
            start_time < end_time
        ),

    CONSTRAINT professional_working_hours_client_fkey
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE,

    CONSTRAINT professional_working_hours_professional_fkey
        FOREIGN KEY (
            professional_id,
            client_id
        )
        REFERENCES professionals(
            id,
            client_id
        )
        ON DELETE CASCADE,

    CONSTRAINT professional_working_hours_unique_range
        UNIQUE (
            client_id,
            professional_id,
            weekday,
            start_time,
            end_time
        )
);

CREATE INDEX idx_professional_working_hours_lookup
    ON professional_working_hours (
        client_id,
        professional_id,
        weekday
    );

COMMIT;
