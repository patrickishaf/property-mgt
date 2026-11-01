# Aventra — property management API

The backend for Aventra, my private app for running my own properties: tenants, leases, rent and installments, payments, messages to tenants, and maintenance. It serves the Aventra web client (dev: `http://localhost:8082`). All money is in a single currency (NGN by default).

## What it does

| Area | Module | Summary |
|---|---|---|
| Accounts & login | `pkg/auth` | Owner sign-up and login for owners, managers and tenants. JWT access + refresh tokens. Invited tenants and managers set their password from an emailed link. |
| Properties | `pkg/property` | Properties and their units, archiving, and assigning property managers. |
| Tenants | `pkg/tenant` | Onboard a tenant together with their first lease, then send an invite link. |
| Leases | `pkg/lease` | Lease lifecycle (active → ended / terminated), optional installment schedules, the due/overdue list, and a daily job that ends expired leases. |
| Payments | `pkg/payment` | Offline payments recorded by owners/managers, online checkout for tenants, allocation across installments, and the dashboard stats. |
| Messaging | `pkg/messaging` | Messages to tenants over in-app, email or SMS, with templates and `{name}` / `{unit}` / `{property}` placeholders. |
| Maintenance | `pkg/maintenance` | Tenants file complaints; owners/managers open issues, link complaints to them and post updates that notify the affected tenants. |

### Who can do what

- **Owner** — sees and manages everything on the properties they own.
- **Property manager** — works for an owner; limited to the properties they are assigned to.
- **Tenant** — sees only their own leases, payments, messages and complaints.

## Stack

Go 1.26 · Gin · GORM on PostgreSQL · Redis · JWT · robfig/cron for background jobs · Air for hot reload.

## Running it locally

### Prerequisites

- Go 1.26+
- PostgreSQL (I use Postgres.app) with an empty database created
- Redis
- [Air](https://github.com/air-verse/air), for hot reload only: `go install github.com/air-verse/air@latest`, with `$(go env GOPATH)/bin` on your `PATH`

### Configuration

Settings are read from a `.env` file in the repo root (it is git-ignored). Keys without a default must be set.

| Variable | Default | Notes |
|---|---|---|
| `APP_ENV` | `development` | `development` turns on the dev-only routes below |
| `APP_NAME` | | |
| `PORT` | | API port, e.g. `3000` |
| `WEB_APP_URL` | | Base URL of the web client; used in invite links |
| `JWT_SECRET_KEY` | | Signing key for all tokens |
| `JWT_AT_EXPIRATION_MINS` | | Access token lifetime |
| `JWT_RT_EXPIRATION_HOURS` | | Refresh token lifetime |
| `PASSWORD_RESET_EXPIRATION_SECONDS` | | |
| `OTP_TIMEOUT_SECONDS` | `180` | |
| `PAYMENT_CURRENCY` | `NGN` | |
| `CHECKOUT_EXPIRY_MINUTES` | `30` | How long an online checkout stays payable |
| `JOB_BACKOFF_TIME_SECONDS` | `10` | Retry backoff for queued email/SMS jobs |
| `MAX_JOB_ATTEMPT_COUNT` | `4` | |
| `WORKER_POOL_CAPACITY` | `100` | |
| `DB_URL` | | Full Postgres URL; if set, the separate `DB_*` values are ignored |
| `DB_HOST` / `DB_PORT` | `localhost` / `5432` | |
| `DB_NAME` / `DB_USER` / `DB_PASSWORD` | | |
| `SSL_MODE` | `disable` | `disable`, `require` or `verify-full` |
| `DB_MAX_OPEN_CONNECTIONS` / `DB_MAX_IDLE_CONNECTIONS` / `DB_MAX_IDLE_CONNECTION_TIME` | — / `45` / `45` | |
| `REDIS_URL` | | Or `REDIS_HOST`, `REDIS_PORT`, `REDIS_USER`, `REDIS_PASSWORD`, `REDIS_DB` |
| `CPAAS_API_BASE_URL` / `CPAAS_API_KEY` | | Email/SMS provider |
| `CPAAS_SENDER_ADDR` / `CPAAS_SENDER_NAME` | / `Aventra` | From address and name on outgoing email |
| `BACHS_API_BASE_URL` / `BACHS_API_KEY` / `BACHS_WH_SECRET` | | Online payment provider |

The schema is created and updated automatically on startup (GORM AutoMigrate in `internal/db/db.go`); there are no migration files to run.

### Commands

```bash
make dev          # run with hot reload (Air)
make run          # run once
make build        # build ./bin/app
make build-linux  # build a linux/amd64 binary in ./bin/linux
make seed         # rebuild the development seed data (needs the API running)
go test ./...     # unit tests
```

### Seed data

`make seed` fills the database with a realistic portfolio that mirrors the web client's demo data, so each screen has something to show while it is wired to the API. Start the API first (`make dev`); the seed runs only when `APP_ENV=development`.

- **Accounts:** one owner, two managers (two properties each) and sixteen tenants, four per property. Every seed account uses an `@seed.aventra.test` email and the password `aventra-dev-123`. The owner is `owner@seed.aventra.test`, and the managers are `tolu@` and `chidinma@`. The run prints every tenant's email and unit at the end.
- **Rent:** annual leases that began nine months before the run date, most split into quarterly installments. The payments leave a mix of paid, partly paid, overdue and upcoming rent. There is also one lease starting in three weeks and one ended lease that still has a balance.
- **Messages and maintenance:** in-app messages from the owner and both managers, plus five complaints. Three issues are linked to some of those complaints, and they have updates.
- **How it works:** the seed calls the running API, so every record goes through the same validation and business rules as the web client's requests. Dates are worked out from the day it runs.
- **Running it again:** each run first deletes everything a previous run created (found by the seed email domain), plus any emails it queued. Your own accounts and data are left alone.

The dataset itself lives in [cmd/seed/data.go](cmd/seed/data.go).

## API documentation

- [docs/API.md](docs/API.md) — the human-readable reference: every endpoint, its request and response, access rules, and which calls each web page makes.
- [docs/openapi.yaml](docs/openapi.yaml) — the same contract as an OpenAPI spec.
- With `APP_ENV=development`, open `http://localhost:<PORT>/docs` for a Swagger UI built from the spec.

Both files are updated in the same change as any change to the API contract.

Every response uses the same envelope:

```json
{ "status": "success", "message": "…", "data": { } }
```

List endpoints return `data` as `{ "items": [], "page": 1, "limit": 10, "total": 0 }`.

### Dev-only routes

These exist only when `APP_ENV=development`:

| Route | Purpose |
|---|---|
| `GET /docs` | Swagger UI |
| `POST /payments/checkout/:reference/complete` | Marks a checkout as paid, standing in for the payment provider's webhook |
| `POST /leases/expiry/run` | Runs the daily lease-expiry job immediately |

## Background jobs

| Job | Schedule | What it does |
|---|---|---|
| Lease expiry (`pkg/lease/expiry.go`) | daily | Ends active leases whose end date has passed |
| cpaas worker (`providers/cpaas/worker.go`) | every 5s | Picks up queued email/SMS jobs |

## Project layout

```
cmd/server/       entrypoint; wires config, repositories, services and routers
config/           env-driven config structs
common/           shared types, errors, JWT, logger, scheduler
internal/db/      connection + AutoMigrate, models, repositories (one model per repository)
internal/net/     server, middleware (JWT, request ID), response helpers
internal/cache/   Redis client
pkg/<module>/     router.go → http_handler.go → service.go, with types.go for requests/responses
providers/        external services: bachs (payments), cpaas (email/SMS queue and worker)
docs/             API.md, openapi.yaml, ENTITIES.md
```

## Not finished yet

- **Sending email/SMS.** Jobs are queued, but the cpaas worker does not process them yet, so invite links and notifications are not delivered. SMS sending is a logging stub.
- **Online payments.** The Bachs client is a stub; use the dev-only complete route to finish a checkout locally.
- **Overview "Export report".** Not built yet.
- **Settings.** `pkg/settings` is a placeholder.
