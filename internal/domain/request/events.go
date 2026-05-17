package request

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeCreated          = "request.created"
	EventTypeSubmitted        = "request.submitted"
	EventTypeReturned         = "request.returned"
	EventTypeResubmitted      = "request.resubmitted"
	EventTypeApproved         = "request.approved"
	EventTypeRejected         = "request.rejected"
	EventTypeRequestCommented = "request.request_commented"
)

type Event struct {
	OrgID      uuid.UUID
	RequestID  uuid.UUID
	Version    int32
	Type       string
	OccurredAt time.Time
	Payload    json.RawMessage
}

// Payload types (MVP-031 minimal)

type CreatedPayload struct {
	Title              string `json:"title"`
	WorkflowTemplateID string `json:"workflow_template_id,omitempty"`
}

type RejectedPayload struct {
	Reason string `json:"reason"`
}

type ReturnedPayload struct {
	Reason string `json:"reason"`
}

type RequestCommentedPayload struct {
	CommentID   string    `json:"comment_id"`
	UserID      string    `json:"user_id"`
	Content     string    `json:"content"`
	CommentedAt time.Time `json:"commented_at"`
}

func decodePayload[T any](raw json.RawMessage) (T, error) {
	var out T
	if len(raw) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("invalid payload: %w", err)
	}
	return out, nil
}

func (p CreatedPayload) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(p.Title) > 200 {
		return fmt.Errorf("title too long (max 200 chars)")
	}
	return nil
}

func (p RejectedPayload) Validate() error {
	if p.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	if len(p.Reason) > 500 {
		return fmt.Errorf("reason too long (max 500 chars)")
	}
	return nil
}

func (p ReturnedPayload) Validate() error {
	if p.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	if len(p.Reason) > 500 {
		return fmt.Errorf("reason too long (max 500 chars)")
	}
	return nil
}

func (p RequestCommentedPayload) Validate() error {
	if p.CommentID == "" {
		return fmt.Errorf("comment_id is required")
	}
	if _, err := uuid.Parse(p.CommentID); err != nil {
		return fmt.Errorf("comment_id must be valid UUID: %w", err)
	}

	if p.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if _, err := uuid.Parse(p.UserID); err != nil {
		return fmt.Errorf("user_id must be valid UUID: %w", err)
	}

	if p.Content == "" {
		return fmt.Errorf("content is required")
	}
	if len(p.Content) > 1000 {
		return fmt.Errorf("content too long (max 1000 chars)")
	}

	if p.CommentedAt.IsZero() {
		return fmt.Errorf("commented_at is required")
	}

	return nil
}
