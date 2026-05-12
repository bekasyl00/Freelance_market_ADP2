# Local Database Quick Start

This is the fastest setup for backend developers. It does not require Supabase.

## Start PostgreSQL

```bash
docker compose up -d postgres
```

Connection string for Go services:

```env
DATABASE_URL=postgresql://freelance:freelance@localhost:5433/freelance_market?sslmode=disable
```

The schema is loaded automatically from:

```text
database/init/001_schema.sql
```

This happens only when the Docker volume is created for the first time.

## Reset Database

Use this when you want to recreate tables from scratch:

```bash
docker compose down -v
docker compose up -d postgres
```

## Connect With psql

```bash
psql "postgresql://freelance:freelance@localhost:5433/freelance_market?sslmode=disable"
```

Or through Docker:

```bash
docker exec -it freelance-postgres psql -U freelance -d freelance_market
```

## Tables

- `users`, `user_skills`, `user_sessions`
- `jobs`, `job_skills`, `proposals`
- `payment_accounts`, `escrows`, `transactions`

## For Go Backend

Each backend service can use the same local `DATABASE_URL` during development.

Recommended ownership:

- user-service: `users`, `user_skills`, `user_sessions`
- job-service: `jobs`, `job_skills`, `proposals`
- payment-service: `payment_accounts`, `escrows`, `transactions`

Supabase can be added later by running the same schema in a hosted Postgres database and changing only `DATABASE_URL`.
