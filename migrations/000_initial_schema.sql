BEGIN;

-- Schema-base histórico do Flitta.
-- Representa o estado efetivo do banco imediatamente antes da migration 001.
--
-- customer_phone e selected_appointment_id já eram adicionados pelo antigo
-- runMigrations() da aplicação e, por isso, fazem parte deste baseline.

CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    name TEXT,
    phone TEXT,
    email VARCHAR(100),
    password VARCHAR(255),
    business_type VARCHAR(50) DEFAULT 'other',

    CONSTRAINT clients_phone_key
        UNIQUE (phone)
);

CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    client_id INTEGER,
    name VARCHAR(100),
    duration INTEGER,

    CONSTRAINT unique_service
        UNIQUE (client_id, name)
);

CREATE TABLE IF NOT EXISTS appointments (
    id SERIAL PRIMARY KEY,
    name TEXT,
    service TEXT,
    date TEXT,
    time TEXT,
    client_id INTEGER,

    customer_phone VARCHAR(64),

    reminder_sent BOOLEAN
        NOT NULL
        DEFAULT FALSE,

    reminder_sent_at TIMESTAMP WITHOUT TIME ZONE,

    -- Restrição histórica anterior ao lifecycle.
    -- A migration 001 remove esta constraint e cria
    -- unique_schedule_active.
    CONSTRAINT unique_shedule
        UNIQUE (client_id, date, time)
);

CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,

    phone VARCHAR(50),
    client_id INTEGER,

    state VARCHAR(50),
    name VARCHAR(100),
    service VARCHAR(100),
    date VARCHAR(50),
    time VARCHAR(50),

    updated_at TIMESTAMP WITHOUT TIME ZONE
        DEFAULT CURRENT_TIMESTAMP,

    selected_appointment_id INTEGER
        NOT NULL
        DEFAULT 0,

    suggested_time TEXT,

    -- Antes do suporte multiempresa, o telefone sozinho
    -- identificava uma sessão.
    CONSTRAINT user_sessions_phone_key
        UNIQUE (phone)
);

CREATE TABLE IF NOT EXISTS working_hours (
    id SERIAL PRIMARY KEY,
    client_id INTEGER,
    start_time VARCHAR(10),
    end_time VARCHAR(10),
    interval_minutes INTEGER
);

COMMIT;