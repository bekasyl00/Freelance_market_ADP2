package natsdelivery

import (
	"context"
	"encoding/json"
	"log"

	"notification_service/internal/models"
	"notification_service/internal/usecase"

	"github.com/nats-io/nats.go"
)

const (
	SubjectApplicationSubmitted = "jobs.application.submitted"
	SubjectPaymentEvents        = "payments.events"
)

type Subscriber struct {
	nc *nats.Conn
	uc *usecase.NotificationUsecase
}

func NewSubscriber(nc *nats.Conn, uc *usecase.NotificationUsecase) *Subscriber {
	return &Subscriber{nc: nc, uc: uc}
}

func (s *Subscriber) Subscribe(ctx context.Context) error {
	if _, err := s.nc.Subscribe(SubjectApplicationSubmitted, func(msg *nats.Msg) {
		var event models.JobApplicationSubmitted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("decode job application event: %v", err)
			return
		}
		if err := s.uc.NotifyJobApplication(ctx, event); err != nil {
			log.Printf("notify job application: %v", err)
		}
	}); err != nil {
		return err
	}

	if _, err := s.nc.Subscribe(SubjectPaymentEvents, func(msg *nats.Msg) {
		var event models.PaymentEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("decode payment event: %v", err)
			return
		}
		if err := s.uc.NotifyPaymentEvent(ctx, event); err != nil {
			log.Printf("notify payment event: %v", err)
		}
	}); err != nil {
		return err
	}

	return s.nc.Flush()
}
