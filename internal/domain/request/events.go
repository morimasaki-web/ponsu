package request

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeCreated   = "request.created"
	EventTypeSubmitted = "request.submitted"
	EventTypeReturned  = "request.returned"
	EventTypeResubmitted = "request.resubmitted"
	EventTypeApproved  = "request.approved"
	EventTypeRejected  = "request.rejected"
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
