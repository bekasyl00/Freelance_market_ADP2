package natsdelivery

import (
	"context"
	"encoding/json"

	"job_service/internal/domain"

	"github.com/nats-io/nats.go"
)

const SubjectApplicationSubmitted = "jobs.application.submitted"

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

func (p *Publisher) PublishApplicationSubmitted(ctx context.Context, e domain.JobApplicationSubmitted) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return p.nc.Publish(SubjectApplicationSubmitted, b)
}
