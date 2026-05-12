# Supabase Postgres Schema

This database is designed for the Go backend only. The Vue frontend should call the Go gateway, not Supabase tables directly.

## Migration Order

Apply migrations in this order because `jobs` depends on `users`, and `payments` depends on both:

```text
services/user-service/migrations/000001_create_users.up.sql
services/job-service/migrations/000001_create_jobs.up.sql
services/payment-service/migrations/000001_create_payments.up.sql
```

Rollback order is the reverse:

```text
services/payment-service/migrations/000001_create_payments.down.sql
services/job-service/migrations/000001_create_jobs.down.sql
services/user-service/migrations/000001_create_users.down.sql
```

## Supabase Setup

Create a Supabase project, open SQL Editor, and run the `*.up.sql` files in the migration order above.

Use the Supabase Postgres connection string in the Go backend:

```env
DATABASE_URL=postgresql://postgres:<password>@db.<project-ref>.supabase.co:5432/postgres?sslmode=require
```

Keep `SUPABASE_SERVICE_ROLE_KEY` only on the backend if you need Supabase admin APIs. Do not expose it in the frontend.

## Main Tables

- `users`: accounts, roles, profile fields, password hash, rating, balance
- `user_skills`: freelancer skills
- `user_sessions`: refresh token storage for backend auth
- `jobs`: client job posts
- `job_skills`: required skills for each job
- `proposals`: freelancer applications
- `payment_accounts`: available and escrow balances
- `escrows`: protected job funds
- `transactions`: deposit, escrow, release, refund, withdrawal history
