# Freelance Market

Frontend and monitoring setup for the AP2 final project.

## Stack

- Vue 3 + Vite frontend
- vue-i18n with English, Russian, and Kazakh translations
- Go backend planned behind `VITE_API_BASE_URL`
- Supabase Postgres planned for persistence
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
