BEGIN;

CREATE TABLE IF NOT EXISTS communication_channels (
    id SERIAL PRIMARY KEY,

    client_id INTEGER NOT NULL,

    provider VARCHAR(50) NOT NULL,
    channel_type VARCHAR(50) NOT NULL,
    external_address VARCHAR(150) NOT NULL,

    active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT communication_channels_client_id_fkey
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE,

    CONSTRAINT communication_channels_provider_not_empty
        CHECK (TRIM(provider) <> ''),

    CONSTRAINT communication_channels_type_not_empty
        CHECK (TRIM(channel_type) <> ''),

    CONSTRAINT communication_channels_address_not_empty
        CHECK (TRIM(external_address) <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS
    idx_communication_channels_external_unique
ON communication_channels (
    provider,
    channel_type,
    external_address
);

CREATE INDEX IF NOT EXISTS
    idx_communication_channels_client_id
ON communication_channels (client_id);

COMMIT;
