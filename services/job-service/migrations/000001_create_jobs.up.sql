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
