package domain

import "context"

type JobStatus string

const (
	JobStatusOpen        JobStatus = "OPEN"
	JobStatusInProgress  JobStatus = "IN_PROGRESS"
	JobStatusCompleted   JobStatus = "COMPLETED"
	JobStatusCancelled   JobStatus = "CANCELLED"
)

type Job struct {
	ID            int64
	ClientID      int64
	Title         string
	Description   string
	BudgetCents   int64
	Status        JobStatus
	CreatedAtUnix int64
}

type JobApplication struct {
	ID           int64
	JobID        int64
	FreelancerID int64
	CoverLetter  string
	BidCents     int64
}

type JobApplicationSubmitted struct {
	JobID           int64  `json:"job_id"`
	FreelancerID    int64  `json:"freelancer_id"`
	ApplicationID   int64  `json:"application_id"`
	CoverLetter     string `json:"cover_letter"`
	BidCents        int64  `json:"bid_cents"`
	CreatedAtUnix   int64  `json:"created_at_unix"`
}

type JobRepository interface {
	CreateJob(ctx context.Context, j *Job) (*Job, error)
	ListOpenJobs(ctx context.Context, limit int32, offset int64) ([]Job, error)
	GetJobByID(ctx context.Context, id int64) (*Job, error)
	CountApplicationsByJobID(ctx context.Context, jobID int64) (int32, error)
	InsertApplication(ctx context.Context, a *JobApplication) (int64, error)
	HasApplication(ctx context.Context, jobID, freelancerID int64) (bool, error)
	CompleteJobByClient(ctx context.Context, jobID, clientID int64) (*Job, error)
}

type ApplicationPublisher interface {
	PublishApplicationSubmitted(ctx context.Context, e JobApplicationSubmitted) error
}

type JobUsecase interface {
	CreateJob(ctx context.Context, clientID int64, title, description string, budgetCents int64) (*Job, error)
	ListJobs(ctx context.Context, pageSize int32, pageToken string) ([]Job, string, error)
	ApplyToJob(ctx context.Context, jobID, freelancerID int64, coverLetter string, bidCents int64) (int64, error)
	GetJobDetails(ctx context.Context, jobID int64) (*Job, int32, error)
	CompleteJob(ctx context.Context, jobID, clientID int64) (*Job, error)
}
