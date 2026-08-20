BEGIN;

-- Toda sessão precisa pertencer a uma empresa e a um telefone.
-- O preflight confirmou que os dados atuais já atendem a essa regra.
ALTER TABLE user_sessions
    ALTER COLUMN phone SET NOT NULL,
    ALTER COLUMN client_id SET NOT NULL;

-- Nova identidade da sessão:
-- uma conversa pertence ao par empresa + telefone do cliente.
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_sessions_client_phone_unique
    ON user_sessions (client_id, phone);

-- Garante integridade entre a sessão e a empresa.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'user_sessions_client_id_fkey'
          AND conrelid = 'user_sessions'::regclass
    ) THEN
        ALTER TABLE user_sessions
            ADD CONSTRAINT user_sessions_client_id_fkey
            FOREIGN KEY (client_id)
            REFERENCES clients(id)
            ON DELETE CASCADE;
    END IF;
END
$$;

COMMIT;