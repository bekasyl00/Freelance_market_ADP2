package usecase

import (
	"context"
	"fmt"
	"os"
	"strings"

	"notification_service/internal/models"
)

type NotificationUsecase struct {
	mailer models.Mailer
}

func NewNotificationUsecase(mailer models.Mailer) *NotificationUsecase {
	return &NotificationUsecase{mailer: mailer}
}

func (u *NotificationUsecase) NotifyJobApplication(ctx context.Context, event models.JobApplicationSubmitted) error {
	to := recipientForIntUser(event.FreelancerID)
	return u.mailer.Send(ctx, models.Email{
		To:      to,
		Subject: "Your job application was submitted",
		Body:    fmt.Sprintf("Application #%d for job #%d was submitted with a bid of %.2f.", event.ApplicationID, event.JobID, float64(event.BidCents)/100),
	})
}

func (u *NotificationUsecase) NotifyPaymentEvent(ctx context.Context, event models.PaymentEvent) error {
	to := recipientForUser(event.UserID)
	return u.mailer.Send(ctx, models.Email{
		To:      to,
		Subject: paymentSubject(event.EventType),
		Body:    fmt.Sprintf("%s for %s %.2f.", paymentSubject(event.EventType), event.Currency, float64(event.AmountCents)/100),
	})
}

func paymentSubject(eventType string) string {
	switch eventType {
	case "payment.deposit.completed":
		return "Deposit completed"
	case "payment.escrow.held":
		return "Escrow payment held"
	case "payment.escrow.released":
		return "Escrow payment released"
	case "payment.escrow.refunded":
		return "Escrow payment refunded"
	default:
		return "Payment update"
	}
}

func recipientForUser(userID string) string {
	if fallback := strings.TrimSpace(os.Getenv("MAIL_TO")); fallback != "" {
		return fallback
	}
	return strings.TrimSpace(userID)
}

func recipientForIntUser(userID int64) string {
	if fallback := strings.TrimSpace(os.Getenv("MAIL_TO")); fallback != "" {
		return fallback
	}
	return fmt.Sprintf("user-%d@example.local", userID)
}
