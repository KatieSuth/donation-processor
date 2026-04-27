### Project Status
[![Backend](https://github.com/KatieSuth/donation-processor/actions/workflows/backend.yml/badge.svg)](https://github.com/KatieSuth/donation-processor/actions)
[![Backend Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/KatieSuth/3a4113f9325b9a4a8215ff2fcd018fbf/raw/donation-processor-coverage.json)](https://github.com/KatieSuth/donation-processor/actions)
[![Frontend](https://github.com/KatieSuth/donation-processor/actions/workflows/frontend.yml/badge.svg)](https://github.com/KatieSuth/donation-processor/actions)

# Donation Processor

Minimal full-stack donation processor with:
- Go + Gin backend
- PostgreSQL + Goose migrations + sqlc
- Next.js frontend
- Docker Compose + Caddy local HTTPS setup

The backend exposes donation ingestion and status management APIs, plus `GET /health`.

## Tech Stack

- Frontend: Next.js (App Router), TypeScript, Tailwind CSS
- Backend: Go, Gin, pgx, Goose, sqlc
- Database: PostgreSQL
- Infra: Docker Compose, Caddy

## Quick Start

### Prerequisites
- Docker + Docker Compose v2
- `make` (optional)

### Env files
Using the provided .env.example files in the `frontend/` and `backend/` directories, create your own .env files with your own values. The Postgres database will be initially created with whatever values you provide it, so ensure you use a secure "POSTGRES_PASSWORD" value.

### Development

```bash
make dev
```

or

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

### Certificates

On a first-time-run, Caddy's local CA cert must be exported from the container and trusted.

1. Fetch the certificate (run from the project root): `docker exec donation-processor-caddy cat /data/caddy/pki/authorities/local/root.crt > caddy-root.crt`
2. Install `caddy-root.crt` into your OS/browser trust store (filepath assumes you're running from the directory where the caddy-root.crt file lives, adjust accordingly):

On Windows:
Run PowerShell as Administrator. 
```
Import-Certificate -FilePath ".\caddy-root.crt" -CertStoreLocation Cert:\LocalMachine\Root
```

On Linux:
```
sudo cp caddy-root.crt /usr/local/share/ca-certificates/caddy.crt
sudo update-ca-certificates
```

On macOS:
```
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain caddy-root.cr
```

Additional step for Firefox, since it uses its own trust store:
```
1. Settings → Privacy & Security → scroll to the bottom → View Certificates
2. Authorities tab → Import
3. Select the caddy-root.crt file
4. Check "Trust this CA to identify websites"
5. OK → restart Firefox
```

## API

Base URL (local): `https://donation-processor.localhost/api`

### Health check

```bash
curl https://donation-processor.localhost/api/health
```

For endpoint definitions and response schemas, use the OpenAPI source of truth:

- `backend/openapi.yaml`

Regarding the 409 Conflict status: any calls to `POST /donations` with a UUID that already exists in the database is considered a conflict, as that is the unique identifier for a record. I considered using some combination of the donor ID, non profit ID, value, and created time as an additional constraint, but ultimately decided against it as this is not a system that creates the donations but rather that receives donations from some other source, as the instructions gave no requirement to submit a donation through the UI; therefore, it is not this system's job to reject a new donation on any basis besides "this unique identifier already exists." If a system is processing duplicate donations with different UUIDs the bug is on that side, not this one, and rejecting the update here could further obfuscate any such problems.

## Database Schema and SQLC

All old domain migrations were removed. The schema now starts from:

- `backend/sql/migrations/001_initial_schema.sql`

This migration creates the `donations` table:

- `donations(uuid, amount, currency, payment_method, nonprofit_id, donor_id, status, created_at, updated_at)`

Donation queries live in:

- `backend/sql/queries/donations.sql`

Regenerate sqlc code after query/schema changes:

```bash
cd backend
sqlc generate -f sql/sqlc.yaml
```

## Environment Variables

### Backend (`backend/.env.example`)

- `PORT` (default `8080`)
- `GIN_MODE` (default `release`)
- `DATABASE_URL` (default `postgres://app:secret@db:5432/appdb`)
- `DATABASE_URL_TESTS` (default `postgres://app:secret@localhost:5432/appdb`)
- `POSTGRES_USER` (default `app`)
- `POSTGRES_PASSWORD` (default `secret`)
- `POSTGRES_DB` (default `appdb`)
- `FRONTEND_URL` (default `https://donation-processor.localhost`, reserved for frontend-origin/CORS config)

### Frontend (`frontend/.env.example`)

- `NEXT_PUBLIC_API_URL` (default `https://donation-processor.localhost/api`)
- `NODE_ENV` (default `development`)

## Useful Commands

```bash
make dev
make dev-down
make health
make test
make test-coverage
```
