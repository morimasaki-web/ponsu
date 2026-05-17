package requests

import (
	"context"

	"github.com/google/uuid"
)

type NotificationKind string

const (
	NotificationKindSubmitted NotificationKind = "submitted"
	NotificationKindApproved  NotificationKind = "approved"
	NotificationKindRejected  NotificationKind = "rejected"
)

type Notification struct {
	Kind         NotificationKind
	OrgID        uuid.UUID
	RequestID    uuid.UUID
	RequestTitle string
	Actor        string
	URL          string
}

type Notifier interface {
	Notify(ctx context.Context, n Notification) error
}
