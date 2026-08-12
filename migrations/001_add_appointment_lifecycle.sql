BEGIN;

-- Adiciona o estado atual de cada agendamento.
-- Os registros antigos receberão automaticamente o status "scheduled".
ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS status VARCHAR(20)
        NOT NULL
        DEFAULT 'scheduled';

-- Adiciona datas para auditoria e análises futuras.
ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITHOUT TIME ZONE
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITHOUT TIME ZONE
        NOT NULL
        DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMP WITHOUT TIME ZONE;

-- Garante que somente status conhecidos possam ser salvos.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'appointments_status_check'
          AND conrelid = 'appointments'::regclass
    ) THEN
        ALTER TABLE appointments
            ADD CONSTRAINT appointments_status_check
            CHECK (
                status IN (
                    'scheduled',
                    'confirmed',
                    'completed',
                    'cancelled',
                    'no_show'
                )
            );
    END IF;
END
$$;

-- Remove a restrição antiga.
-- O nome original contém o erro de escrita "shedule".
ALTER TABLE appointments
    DROP CONSTRAINT IF EXISTS unique_shedule;

-- Proteção para bancos onde o nome possa estar corrigido.
ALTER TABLE appointments
    DROP CONSTRAINT IF EXISTS unique_schedule;

-- Um horário continua exclusivo enquanto o agendamento estiver ativo.
-- Um horário cancelado poderá ser agendado novamente.
CREATE UNIQUE INDEX IF NOT EXISTS unique_schedule_active
    ON appointments (client_id, date, time)
    WHERE status <> 'cancelled';

-- Índices que facilitarão as consultas dos indicadores.
CREATE INDEX IF NOT EXISTS idx_appointments_client_status
    ON appointments (client_id, status);

CREATE INDEX IF NOT EXISTS idx_appointments_client_date
    ON appointments (client_id, date);

COMMIT;