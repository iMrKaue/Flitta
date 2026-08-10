BEGIN;

-- Preço atual do serviço.
ALTER TABLE services
    ADD COLUMN IF NOT EXISTS price NUMERIC(10, 2)
        NOT NULL
        DEFAULT 0.00;

-- Valor preservado no momento do agendamento.
ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS price_snapshot NUMERIC(10, 2)
        NOT NULL
        DEFAULT 0.00;

-- A duração começa como nullable para permitir o backfill
-- seguro dos agendamentos existentes.
ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS duration_snapshot INTEGER;

-- Para agendamentos antigos, recupera a duração atual do serviço
-- somente quando ainda não existe snapshot.
UPDATE appointments AS a
SET duration_snapshot = COALESCE(s.duration, 30)
FROM services AS s
WHERE s.client_id = a.client_id
  AND s.name = a.service
  AND a.duration_snapshot IS NULL;

-- Caso algum agendamento antigo não encontre serviço correspondente,
-- utiliza 30 minutos como fallback.
UPDATE appointments
SET duration_snapshot = 30
WHERE duration_snapshot IS NULL;

-- Novos registros terão 30 minutos como fallback,
-- mas o código Go preencherá a duração real do serviço.
ALTER TABLE appointments
    ALTER COLUMN duration_snapshot SET DEFAULT 30,
    ALTER COLUMN duration_snapshot SET NOT NULL;

-- Impede preços negativos nos serviços.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'services_price_non_negative'
          AND conrelid = 'services'::regclass
    ) THEN
        ALTER TABLE services
            ADD CONSTRAINT services_price_non_negative
            CHECK (price >= 0);
    END IF;
END
$$;

-- Impede valores históricos negativos.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'appointments_price_snapshot_non_negative'
          AND conrelid = 'appointments'::regclass
    ) THEN
        ALTER TABLE appointments
            ADD CONSTRAINT appointments_price_snapshot_non_negative
            CHECK (price_snapshot >= 0);
    END IF;
END
$$;

-- Uma duração precisa ser maior que zero.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'appointments_duration_snapshot_positive'
          AND conrelid = 'appointments'::regclass
    ) THEN
        ALTER TABLE appointments
            ADD CONSTRAINT appointments_duration_snapshot_positive
            CHECK (duration_snapshot > 0);
    END IF;
END
$$;

COMMIT;