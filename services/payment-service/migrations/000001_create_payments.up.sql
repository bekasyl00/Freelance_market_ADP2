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
