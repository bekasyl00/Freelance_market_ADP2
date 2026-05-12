-- Local PostgreSQL schema for the Go backend.
-- The Vue frontend should call the Go gateway and should not query these tables directly.

create extension if not exists pgcrypto;

create table if not exists users (
  id uuid primary key default gen_random_uuid(),
  email text not null unique,
  password_hash text not null,
  full_name text not null,
  role text not null check (role in ('client', 'freelancer', 'admin')),
  avatar_url text,
  bio text,
  rating numeric(3, 2) not null default 0 check (rating >= 0 and rating <= 5),
  completed_jobs integer not null default 0 check (completed_jobs >= 0),
  is_verified boolean not null default false,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists user_skills (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  skill text not null,
  created_at timestamptz not null default now(),
  unique (user_id, skill)
);

create table if not exists user_sessions (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  refresh_token_hash text not null unique,
  expires_at timestamptz not null,
  revoked_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists idx_users_role on users(role);
create index if not exists idx_users_created_at on users(created_at desc);
create index if not exists idx_user_skills_skill on user_skills(lower(skill));
create index if not exists idx_user_sessions_user_id on user_sessions(user_id);

create or replace function set_updated_at()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

drop trigger if exists trg_users_updated_at on users;
create trigger trg_users_updated_at
before update on users
for each row execute function set_updated_at();

create table if not exists jobs (
  id uuid primary key default gen_random_uuid(),
  client_id uuid not null references users(id) on delete restrict,
  title text not null,
  description text not null,
  budget_cents bigint not null check (budget_cents > 0),
  currency char(3) not null default 'USD',
  status text not null default 'open' check (status in ('open', 'in_progress', 'completed', 'cancelled')),
  deadline date,
  selected_freelancer_id uuid references users(id) on delete set null,
  completed_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint selected_freelancer_is_not_client check (
    selected_freelancer_id is null or selected_freelancer_id <> client_id
  )
);

create table if not exists job_skills (
  id uuid primary key default gen_random_uuid(),
  job_id uuid not null references jobs(id) on delete cascade,
  skill text not null,
  created_at timestamptz not null default now(),
  unique (job_id, skill)
);

create table if not exists proposals (
  id uuid primary key default gen_random_uuid(),
  job_id uuid not null references jobs(id) on delete cascade,
  freelancer_id uuid not null references users(id) on delete cascade,
  cover_letter text not null,
  bid_cents bigint not null check (bid_cents > 0),
  estimated_days integer not null check (estimated_days > 0),
  status text not null default 'pending' check (status in ('pending', 'accepted', 'rejected', 'withdrawn')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (job_id, freelancer_id)
);

create index if not exists idx_jobs_client_id on jobs(client_id);
create index if not exists idx_jobs_selected_freelancer_id on jobs(selected_freelancer_id);
create index if not exists idx_jobs_status_created_at on jobs(status, created_at desc);
create index if not exists idx_jobs_deadline on jobs(deadline);
create index if not exists idx_job_skills_skill on job_skills(lower(skill));
create index if not exists idx_proposals_job_id on proposals(job_id);
create index if not exists idx_proposals_freelancer_id on proposals(freelancer_id);
create index if not exists idx_proposals_status on proposals(status);

drop trigger if exists trg_jobs_updated_at on jobs;
create trigger trg_jobs_updated_at
before update on jobs
for each row execute function set_updated_at();

drop trigger if exists trg_proposals_updated_at on proposals;
create trigger trg_proposals_updated_at
before update on proposals
for each row execute function set_updated_at();

create table if not exists payment_accounts (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null unique references users(id) on delete cascade,
  available_cents bigint not null default 0 check (available_cents >= 0),
  escrow_cents bigint not null default 0 check (escrow_cents >= 0),
  currency char(3) not null default 'USD',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists escrows (
  id uuid primary key default gen_random_uuid(),
  job_id uuid not null unique references jobs(id) on delete restrict,
  client_id uuid not null references users(id) on delete restrict,
  freelancer_id uuid not null references users(id) on delete restrict,
  amount_cents bigint not null check (amount_cents > 0),
  currency char(3) not null default 'USD',
  status text not null default 'held' check (status in ('held', 'released', 'refunded', 'cancelled')),
  held_at timestamptz not null default now(),
  released_at timestamptz,
  refunded_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint escrow_users_are_different check (client_id <> freelancer_id)
);

create table if not exists transactions (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete restrict,
  job_id uuid references jobs(id) on delete set null,
  escrow_id uuid references escrows(id) on delete set null,
  type text not null check (type in ('deposit', 'escrow_hold', 'escrow_release', 'refund', 'withdrawal')),
  amount_cents bigint not null check (amount_cents > 0),
  currency char(3) not null default 'USD',
  status text not null default 'pending' check (status in ('pending', 'completed', 'failed', 'cancelled')),
  provider text,
  provider_reference text,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_payment_accounts_user_id on payment_accounts(user_id);
create index if not exists idx_escrows_client_id on escrows(client_id);
create index if not exists idx_escrows_freelancer_id on escrows(freelancer_id);
create index if not exists idx_escrows_status on escrows(status);
create index if not exists idx_transactions_user_created_at on transactions(user_id, created_at desc);
create index if not exists idx_transactions_job_id on transactions(job_id);
create index if not exists idx_transactions_escrow_id on transactions(escrow_id);
create index if not exists idx_transactions_status on transactions(status);

drop trigger if exists trg_payment_accounts_updated_at on payment_accounts;
create trigger trg_payment_accounts_updated_at
before update on payment_accounts
for each row execute function set_updated_at();

drop trigger if exists trg_escrows_updated_at on escrows;
create trigger trg_escrows_updated_at
before update on escrows
for each row execute function set_updated_at();
