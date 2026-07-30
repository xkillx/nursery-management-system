package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	emaildomain "nursery-management-system/api/internal/modules/email/domain"
)

type EmailAdapter struct {
	enqueuer emaildomain.EmailEnqueuer
}

func NewEmailAdapter(enqueuer emaildomain.EmailEnqueuer) *EmailAdapter {
	return &EmailAdapter{enqueuer: enqueuer}
}

func (a *EmailAdapter) SendPasswordReset(ctx context.Context, tenantID, branchID uuid.UUID, toEmail string, resetURL string) error {
	payloadJSON, err := json.Marshal(map[string]string{
		"reset_url": resetURL,
	})
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	_, err = a.enqueuer.Enqueue(ctx, tenantID, branchID, emaildomain.EnqueueParams{
		EventType:       "password_reset",
		Recipient:       toEmail,
		Subject:         "Reset your password",
		TemplateName:    "password_reset",
		TemplateVersion: 1,
		PayloadJSON:     payloadJSON,
		EntityID:        toEmail,
	})
	return err
}
