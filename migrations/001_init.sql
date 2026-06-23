-- accounts: holds user balances
CREATE TABLE IF NOT EXISTS accounts (
    id         BIGSERIAL PRIMARY KEY,
    owner_id   BIGINT        NOT NULL,
    balance    NUMERIC(19,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now()
);

-- transfers: idempotency log for money movements
CREATE TABLE IF NOT EXISTS transfers (
    idempotency_key TEXT          PRIMARY KEY,
    from_id         BIGINT        NOT NULL REFERENCES accounts(id),
    to_id           BIGINT        NOT NULL REFERENCES accounts(id),
    amount          NUMERIC(19,4) NOT NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT now()
);

-- transactions: individual debit/credit entries per user
CREATE TABLE IF NOT EXISTS transactions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT        NOT NULL,
    amount     NUMERIC(19,4) NOT NULL,
    type       TEXT          NOT NULL CHECK (type IN ('credit', 'debit')),
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id, created_at DESC);

-- rate_limit_log: sliding window request tracking, no Redis needed
CREATE TABLE IF NOT EXISTS rate_limit_log (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rate_limit_log_user_time ON rate_limit_log(user_id, created_at);
