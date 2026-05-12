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
