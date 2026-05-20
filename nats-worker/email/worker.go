package email

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

// EmailRequest defines the structure for an incoming email processing request.
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Worker processes NATS messages to send emails in the background.
type Worker struct {
	nc      *nats.Conn
	service Service
}

// NewWorker creates a new Worker instance.
func NewWorker(nc *nats.Conn, service Service) *Worker {
	return &Worker{
		nc:      nc,
		service: service,
	}
}

// HandleMessage is the callback for NATS subscriptions.
func (w *Worker) HandleMessage(m *nats.Msg) {
	var req EmailRequest
	if err := json.Unmarshal(m.Data, &req); err != nil {
		log.Printf("Email Worker: error unmarshaling message: %v", err)
		return
	}

	log.Printf("Email Worker: processing email to %s", req.To)

	ctx := context.Background()
	if err := w.service.SendEmail(ctx, req.To, req.Subject, req.Body); err != nil {
		log.Printf("Email Worker: error sending email: %v", err)
		return
	}

	log.Printf("Email Worker: email sent successfully to %s", req.To)
}
