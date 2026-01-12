package request

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
)

type Request struct {
	OrgID uuid.UUID
	ID    uuid.UUID

	Title  string
	Status Status

	CreatedAt   time.Time
	SubmittedAt *time.Time
	DecidedAt   *time.Time
}
