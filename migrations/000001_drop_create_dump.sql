CREATE TABLE IF NOT EXISTS wallets (
    uuid UUID PRIMARY KEY,
    balance DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

TRUNCATE TABLE wallets;


INSERT INTO wallets (uuid, balance, currency, created_at) VALUES
    (gen_random_uuid(), 0.00, 'USD', now() - interval '5 days'),
    (gen_random_uuid(), 10.25, 'EUR', now() - interval '4 days'),
    ('e3e0fde1-a2d9-4953-8898-5bb6f5ea1bab', 50.81, 'GBP', now() - interval '3 days'),
    (gen_random_uuid(), 176.21, 'USD', now() - interval '2 days'),
    (gen_random_uuid(), 82939.32, 'JPY', now() - interval '1 day');