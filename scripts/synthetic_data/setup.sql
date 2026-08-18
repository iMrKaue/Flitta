BEGIN;

UPDATE services
SET duration = 60,
    price = 55.00
WHERE client_id = 19
    AND name = 'Corte';

UPDATE services
SET duration = 60,
    price = 70.00
WHERE client_id = 19
    AND name = 'Escova';

UPDATE services
SET duration = 180,
    price = 250.00
WHERE client_id = 19
    AND name = 'Progressiva';

COMMIT;
