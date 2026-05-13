-- Таблицы для Job Service (применить в той же БД, что и users — согласовать с Бекасылом).
-- Не трогает payment и user-схему.

CREATE TABLE IF NOT EXISTS jobs (
    id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    budget_cents BIGINT NOT NULL DEFAULT 0 CHECK (budget_cents >= 0),
    status TEXT NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_created ON jobs (status, created_at DESC);

CREATE TABLE IF NOT EXISTS job_applications (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    freelancer_id BIGINT NOT NULL,
    cover_letter TEXT NOT NULL DEFAULT '',
    bid_cents BIGINT NOT NULL DEFAULT 0 CHECK (bid_cents >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (job_id, freelancer_id)
);

CREATE INDEX IF NOT EXISTS idx_job_applications_job ON job_applications (job_id);
