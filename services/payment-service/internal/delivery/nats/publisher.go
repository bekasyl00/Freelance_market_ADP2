package natsdelivery

import (
	"context"
	"encoding/json"

	"payment_service/internal/models"

	"github.com/nats-io/nats.go"
)

const SubjectPaymentEvents = "payments.events"

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

func (p *Publisher) PublishPaymentEvent(ctx context.Context, e models.PaymentEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return p.nc.Publish(SubjectPaymentEvents, b)
}
