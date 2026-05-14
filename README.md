# Freelance Market

Frontend and monitoring setup for the AP2 final project.

## Stack

- Vue 3 + Vite frontend
- vue-i18n with English, Russian, and Kazakh translations
- Go backend planned behind `VITE_API_BASE_URL`
- Local PostgreSQL for backend development, Supabase Postgres later for hosted deployment
- Prometheus metrics collection
- Grafana dashboards and alert rules

## Run Frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`.

## Build Frontend

```bash
cd frontend
npm run build
```

The Vercel config is in `frontend/vercel.json`. Set `VITE_API_BASE_URL` in Vercel when the Go gateway is deployed.

## Local Database

For the backend team, use local PostgreSQL first. Supabase can be connected later with the same schema.

Start the database:

```bash
docker compose up -d postgres
```

Use this connection string in Go services:

```env
DATABASE_URL=postgresql://freelance:freelance@localhost:5433/freelance_market?sslmode=disable
```

The schema is loaded automatically from:

```text
database/init/001_schema.sql
```

## Run Gateway API

Start PostgreSQL and the REST gateway used by the Vue frontend:

```bash
docker compose up -d postgres gateway
```

Gateway URL:

```text
http://localhost:8088
```

Main frontend API variable:

```env
VITE_API_BASE_URL=http://localhost:8088/api
```

Useful checks:

```bash
curl http://localhost:8088/health
curl http://localhost:8088/api/jobs
curl http://localhost:8088/api/summary
```

API contract for frontend/backend collaboration:

```text
docs/API.md
```

If you need to reset the database:

```bash
docker compose down -v
docker compose up -d postgres
```

Service migration files are still available for Go migration tools:

```text
services/user-service/migrations/000001_create_users.up.sql
services/job-service/migrations/000001_create_jobs.up.sql
services/payment-service/migrations/000001_create_payments.up.sql
```

For future Supabase deployment, use the hosted Postgres connection string:

```env
DATABASE_URL=postgresql://postgres:<password>@db.<project-ref>.supabase.co:5432/postgres?sslmode=require
```

## Run Monitoring

```bash
docker compose up prometheus grafana
```

- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3001`
- Grafana login: `admin` / `admin`

Prometheus is configured to scrape:

- `host.docker.internal:8080` gateway
- `host.docker.internal:8081` user-service
- `host.docker.internal:8082` job-service
- `host.docker.internal:8083` payment-service

The Grafana dashboard is provisioned as `freelance-market-overview`, with alert rules for gateway availability and high 5xx error rate.
