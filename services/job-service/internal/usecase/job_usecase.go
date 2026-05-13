package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"job_service/internal/domain"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type jobUsecase struct {
	repo   domain.JobRepository
	pub    domain.ApplicationPublisher
}

func NewJobUsecase(repo domain.JobRepository, pub domain.ApplicationPublisher) domain.JobUsecase {
	if pub == nil {
		pub = noopPublisher{}
	}
	return &jobUsecase{repo: repo, pub: pub}
}

type noopPublisher struct{}

func (noopPublisher) PublishApplicationSubmitted(context.Context, domain.JobApplicationSubmitted) error {
	return nil
}

func (u *jobUsecase) CreateJob(ctx context.Context, clientID int64, title, description string, budgetCents int64) (*domain.Job, error) {
	title = strings.TrimSpace(title)
	if clientID <= 0 || title == "" {
		return nil, domain.ErrInvalidInput
	}
	if budgetCents < 0 {
		return nil, domain.ErrInvalidInput
	}

	j := &domain.Job{
		ClientID:    clientID,
		Title:       title,
		Description: strings.TrimSpace(description),
		BudgetCents: budgetCents,
		Status:      domain.JobStatusOpen,
	}
	return u.repo.CreateJob(ctx, j)
}

func (u *jobUsecase) ListJobs(ctx context.Context, pageSize int32, pageToken string) ([]domain.Job, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	offset, err := parsePageToken(pageToken)
	if err != nil {
		return nil, "", domain.ErrInvalidInput
	}

	jobs, err := u.repo.ListOpenJobs(ctx, pageSize, offset)
	if err != nil {
		return nil, "", err
	}
	next := formatPageToken(offset, pageSize, len(jobs))
	return jobs, next, nil
}

func (u *jobUsecase) ApplyToJob(ctx context.Context, jobID, freelancerID int64, coverLetter string, bidCents int64) (int64, error) {
	if jobID <= 0 || freelancerID <= 0 {
		return 0, domain.ErrInvalidInput
	}
	coverLetter = strings.TrimSpace(coverLetter)
	if coverLetter == "" {
		return 0, domain.ErrInvalidInput
	}
	if bidCents < 0 {
		return 0, domain.ErrInvalidInput
	}

	job, err := u.repo.GetJobByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrJobNotFound
		}
		return 0, err
	}
	if job.Status != domain.JobStatusOpen {
		return 0, domain.ErrJobNotOpen
	}

	dup, err := u.repo.HasApplication(ctx, jobID, freelancerID)
	if err != nil {
		return 0, err
	}
	if dup {
		return 0, domain.ErrDuplicateProposal
	}

	app := &domain.JobApplication{
		JobID:        jobID,
		FreelancerID: freelancerID,
		CoverLetter:  coverLetter,
		BidCents:     bidCents,
	}
	id, err := u.repo.InsertApplication(ctx, app)
	if err != nil {
		return 0, err
	}

	ev := domain.JobApplicationSubmitted{
		JobID:           jobID,
		FreelancerID:    freelancerID,
		ApplicationID:   id,
		CoverLetter:     coverLetter,
		BidCents:        bidCents,
		CreatedAtUnix:   time.Now().Unix(),
	}
	if err := u.pub.PublishApplicationSubmitted(ctx, ev); err != nil {
		return 0, err
	}
	return id, nil
}

func (u *jobUsecase) GetJobDetails(ctx context.Context, jobID int64) (*domain.Job, int32, error) {
	if jobID <= 0 {
		return nil, 0, domain.ErrInvalidInput
	}
	job, err := u.repo.GetJobByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, domain.ErrJobNotFound
		}
		return nil, 0, err
	}
	n, err := u.repo.CountApplicationsByJobID(ctx, jobID)
	if err != nil {
		return nil, 0, err
	}
	return job, n, nil
}

func (u *jobUsecase) CompleteJob(ctx context.Context, jobID, clientID int64) (*domain.Job, error) {
	if jobID <= 0 || clientID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	job, err := u.repo.GetJobByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrJobNotFound
		}
		return nil, err
	}
	if job.ClientID != clientID {
		return nil, domain.ErrForbiddenClient
	}
	if job.Status == domain.JobStatusCompleted || job.Status == domain.JobStatusCancelled {
		return nil, domain.ErrJobAlreadyTerminal
	}

	out, err := u.repo.CompleteJobByClient(ctx, jobID, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrJobAlreadyTerminal
		}
		return nil, err
	}
	return out, nil
}

func parsePageToken(token string) (int64, error) {
	if token == "" {
		return 0, nil
	}
	return strconv.ParseInt(token, 10, 64)
}

func formatPageToken(offset int64, pageSize int32, got int) string {
	if got < int(pageSize) {
		return ""
	}
	return strconv.FormatInt(offset+int64(pageSize), 10)
}
