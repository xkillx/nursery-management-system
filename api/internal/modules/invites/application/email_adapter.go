package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	emaildomain "nursery-management-system/api/internal/modules/email/domain"
)

type InviteEmailAdapter struct {
	enqueuer emaildomain.EmailEnqueuer
}

func NewInviteEmailAdapter(enqueuer emaildomain.EmailEnqueuer) *InviteEmailAdapter {
	return &InviteEmailAdapter{enqueuer: enqueuer}
}

func (a *InviteEmailAdapter) SendInvite(ctx context.Context, tenantID, branchID uuid.UUID, toEmail, role, acceptURL string) error {
	payloadJSON, err := json.Marshal(map[string]string{
		"role":       role,
		"accept_url": acceptURL,
	})
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	_, err = a.enqueuer.Enqueue(ctx, tenantID, branchID, emaildomain.EnqueueParams{
		EventType:       "invite",
		Recipient:       toEmail,
		Subject:         fmt.Sprintf("You're invited to join as %s", role),
		TemplateName:    "invite",
		TemplateVersion: 1,
		PayloadJSON:     payloadJSON,
		EntityID:        toEmail,
	})
	return err
}
