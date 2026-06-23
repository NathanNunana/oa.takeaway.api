# OA Takeway Financial API

Go REST API built around 6 financial engineering problems — idempotent transfers, CSV processing, rate limiting, IDOR security, audit logging, and ledger reconciliation.

Stack: Go 1.22, Fiber, PostgreSQL.

## Setup

```bash
docker compose up -d   # starts postgres, runs migrations and seed
go run .               # starts server on :8080
```

Copy `.env.example` to `.env` if needed. `DATABASE_URL` is the only required var.

## Env

| Var | Default | |
|-----|---------|--|
| `DATABASE_URL` | — | required |
| `CSV_PATH` | `data/transactions.csv` | |
| `PORT` | `8080` | |

## Endpoints

Protected routes need `X-User-ID: {id}` header.

| Method | Path | Auth | What it does |
|--------|------|------|--------------|
| GET | `/health` | — | DB ping, uptime, memory |
| GET | `/api/accounts` | — | current account balances |
| GET | `/api/balances` | — | per-user balances from CSV |
| GET | `/api/reconcile` | — | ledger vs processor diff |
| POST | `/api/transfer` | ✓ | move money between accounts |
| POST | `/api/pay` | ✓ | rate-limited payment (5/min) |
| GET | `/api/transactions` | ✓ | your transaction history |

## Running the demo

```bash
# starting balances: account 1 = $1000, account 2 = $500
curl :8080/api/accounts

# transfer $50 — call it twice, money only moves once
curl -X POST :8080/api/transfer \
  -H "X-User-ID: 1" -H "Content-Type: application/json" \
  -d '{"from_id":1,"to_id":2,"amount":50,"idempotency_key":"demo-1"}'

# balances and history should reflect the transfer
curl :8080/api/accounts
curl :8080/api/transactions -H "X-User-ID: 1"

# IDOR — ?userId=2 is ignored, you only ever get your own data
curl ":8080/api/transactions?userId=2" -H "X-User-ID: 1"

# rate limiter — 6th request in under a minute returns 429
for i in {1..6}; do
  curl -s -X POST :8080/api/pay \
    -H "X-User-ID: 1" -H "Content-Type: application/json" \
    -d '{"amount":10}' | jq .
done

# redaction — check server logs, password and cardNumber show as [REDACTED]
curl -X POST :8080/api/transfer \
  -H "X-User-ID: 1" -H "Content-Type: application/json" \
  -d '{"from_id":1,"to_id":2,"amount":10,"idempotency_key":"demo-2","password":"secret","cardNumber":"4111111111111111"}'

# csv balances and reconciliation
curl :8080/api/balances
curl :8080/api/reconcile
```

## Structure

```
├── main.go
├── config/          env loading
├── api/
│   ├── handler.go   route handlers
│   ├── middleware.go logging + auth
│   └── health.go    health check
├── transfer/        idempotent transfer
├── csvprocessor/    streaming csv
├── ratelimit/       sql rate limiter
├── reconcile/       ledger diff
├── migrations/      schema + seed
└── data/
    └── transactions.csv
```
