-- Truncate in dependency order so we can re-run safely
TRUNCATE transfers, transactions, rate_limit_log, accounts RESTART IDENTITY CASCADE;

-- Seed two accounts with known IDs
INSERT INTO accounts (id, owner_id, balance) VALUES
    (1, 1, 1000.00),
    (2, 2,  500.00);

-- Advance sequence past the manual IDs
SELECT setval('accounts_id_seq', 2);

-- Seed transaction history for both users
INSERT INTO transactions (user_id, amount, type) VALUES
    (1, 200.00, 'credit'),
    (1,  50.00, 'debit'),
    (1, 300.00, 'credit'),
    (2, 100.00, 'credit'),
    (2,  25.00, 'debit');
