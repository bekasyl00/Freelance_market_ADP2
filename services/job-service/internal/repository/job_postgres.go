package repository

import (
	"context"
	"database/sql"
	"time"

	"job_service/internal/domain"
)

type jobPostgresRepo struct {
	db *sql.DB
}

func NewJobPostgresRepo(db *sql.DB) domain.JobRepository {
	return &jobPostgresRepo{db: db}
}

func (r *jobPostgresRepo) CreateJob(ctx context.Context, j *domain.Job) (*domain.Job, error) {
	const q = `
INSERT INTO jobs (client_id, title, description, budget_cents, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, client_id, title, description, budget_cents, status, created_at`

	var created time.Time
	out := &domain.Job{}
	var status string
	err := r.db.QueryRowContext(ctx, q,
		j.ClientID, j.Title, j.Description, j.BudgetCents, string(j.Status),
	).Scan(&out.ID, &out.ClientID, &out.Title, &out.Description, &out.BudgetCents, &status, &created)
	if err != nil {
		return nil, err
	}
	out.Status = domain.JobStatus(status)
	out.CreatedAtUnix = created.Unix()
	return out, nil
}

func (r *jobPostgresRepo) ListOpenJobs(ctx context.Context, limit int32, offset int64) ([]domain.Job, error) {
	const q = `
SELECT id, client_id, title, description, budget_cents, status, created_at
FROM jobs
WHERE status = 'OPEN'
ORDER BY id DESC
LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *j)
	}
	return list, rows.Err()
}

func (r *jobPostgresRepo) GetJobByID(ctx context.Context, id int64) (*domain.Job, error) {
	const q = `
SELECT id, client_id, title, description, budget_cents, status, created_at
FROM jobs WHERE id = $1`

	row := r.db.QueryRowContext(ctx, q, id)
	j, err := scanJobRow(row)
	return j, err
}

func (r *jobPostgresRepo) CountApplicationsByJobID(ctx context.Context, jobID int64) (int32, error) {
	const q = `SELECT COUNT(*) FROM job_applications WHERE job_id = $1`
	var n int64
	if err := r.db.QueryRowContext(ctx, q, jobID).Scan(&n); err != nil {
		return 0, err
	}
	if n > 1<<31-1 {
		return 1<<31 - 1, nil
	}
	return int32(n), nil
}

func (r *jobPostgresRepo) InsertApplication(ctx context.Context, a *domain.JobApplication) (int64, error) {
	const q = `
INSERT INTO job_applications (job_id, freelancer_id, cover_letter, bid_cents)
VALUES ($1, $2, $3, $4)
RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, q, a.JobID, a.FreelancerID, a.CoverLetter, a.BidCents).Scan(&id)
	return id, err
}

func (r *jobPostgresRepo) HasApplication(ctx context.Context, jobID, freelancerID int64) (bool, error) {
	const q = `
SELECT EXISTS(SELECT 1 FROM job_applications WHERE job_id = $1 AND freelancer_id = $2)`

	var ok bool
	if err := r.db.QueryRowContext(ctx, q, jobID, freelancerID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *jobPostgresRepo) CompleteJobByClient(ctx context.Context, jobID, clientID int64) (*domain.Job, error) {
	const q = `
UPDATE jobs
SET status = 'COMPLETED'
WHERE id = $1 AND client_id = $2
  AND status NOT IN ('COMPLETED', 'CANCELLED')
RETURNING id, client_id, title, description, budget_cents, status, created_at`

	row := r.db.QueryRowContext(ctx, q, jobID, clientID)
	return scanJobRow(row)
}

func scanJob(rows *sql.Rows) (*domain.Job, error) {
	var (
		id            int64
		clientID      int64
		title         string
		description   string
		budget        int64
		status        string
		created       time.Time
	)
	if err := rows.Scan(&id, &clientID, &title, &description, &budget, &status, &created); err != nil {
		return nil, err
	}
	return &domain.Job{
		ID:            id,
		ClientID:      clientID,
		Title:         title,
		Description:   description,
		BudgetCents:   budget,
		Status:        domain.JobStatus(status),
		CreatedAtUnix: created.Unix(),
	}, nil
}

func scanJobRow(row *sql.Row) (*domain.Job, error) {
	var (
		id            int64
		clientID      int64
		title         string
		description   string
		budget        int64
		status        string
		created       time.Time
	)
	if err := row.Scan(&id, &clientID, &title, &description, &budget, &status, &created); err != nil {
		return nil, err
	}
	return &domain.Job{
		ID:            id,
		ClientID:      clientID,
		Title:         title,
		Description:   description,
		BudgetCents:   budget,
		Status:        domain.JobStatus(status),
		CreatedAtUnix: created.Unix(),
	}, nil
}
