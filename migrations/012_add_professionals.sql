BEGIN;

-- Permite que as relações abaixo garantam que
-- serviço e profissional pertencem ao mesmo estabelecimento.
ALTER TABLE services
    ADD CONSTRAINT services_id_client_id_unique
    UNIQUE (id, client_id);

CREATE TABLE professionals (
    id SERIAL PRIMARY KEY,

    client_id INTEGER NOT NULL,

    name VARCHAR(150) NOT NULL,

    active BOOLEAN
        NOT NULL
        DEFAULT TRUE,

    created_at TIMESTAMP
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT professionals_client_id_fkey
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE,

    CONSTRAINT professionals_name_not_blank
        CHECK (BTRIM(name) <> ''),

    CONSTRAINT professionals_id_client_id_unique
        UNIQUE (id, client_id)
);

CREATE UNIQUE INDEX professionals_client_name_normalized_unique
    ON professionals (
        client_id,
        LOWER(BTRIM(name))
    );

CREATE INDEX idx_professionals_client_active
    ON professionals (
        client_id,
        active
    );

CREATE TABLE professional_services (
    client_id INTEGER NOT NULL,

    professional_id INTEGER NOT NULL,

    service_id INTEGER NOT NULL,

    created_at TIMESTAMP
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT professional_services_pkey
        PRIMARY KEY (
            professional_id,
            service_id
        ),

    CONSTRAINT professional_services_client_id_fkey
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE,

    CONSTRAINT professional_services_professional_fkey
        FOREIGN KEY (
            professional_id,
            client_id
        )
        REFERENCES professionals(
            id,
            client_id
        )
        ON DELETE CASCADE,

    CONSTRAINT professional_services_service_fkey
        FOREIGN KEY (
            service_id,
            client_id
        )
        REFERENCES services(
            id,
            client_id
        )
        ON DELETE CASCADE
);

CREATE INDEX idx_professional_services_client_service
    ON professional_services (
        client_id,
        service_id
    );

COMMIT;