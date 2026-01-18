package request

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAggregate_Replay_HappyPath(t *testing.T) {
	orgID := uuid.New()
	reqID := uuid.New()
	base := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	createdPayload, _ := json.Marshal(CreatedPayload{Title: "Demo"})

	agg := NewAggregate()
	err := agg.Replay([]Event{
		{OrgID: orgID, RequestID: reqID, Version: 1, Type: EventTypeCreated, OccurredAt: base.Add(1 * time.Second), Payload: createdPayload},
		{OrgID: orgID, RequestID: reqID, Version: 2, Type: EventTypeSubmitted, OccurredAt: base.Add(2 * time.Second)},
		{OrgID: orgID, RequestID: reqID, Version: 3, Type: EventTypeApproved, OccurredAt: base.Add(3 * time.Second)},
	})
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}

	if agg.Request.OrgID != orgID {
		t.Fatalf("OrgID = %v", agg.Request.OrgID)
	}
	if agg.Request.ID != reqID {
		t.Fatalf("ID = %v", agg.Request.ID)
	}
	if agg.Request.Title != "Demo" {
		t.Fatalf("Title = %q", agg.Request.Title)
	}
	if agg.Request.Status != StatusApproved {
		t.Fatalf("Status = %q", agg.Request.Status)
	}
	if agg.Request.SubmittedAt == nil {
		t.Fatalf("SubmittedAt is nil")
	}
	if agg.Request.DecidedAt == nil {
		t.Fatalf("DecidedAt is nil")
	}
	if agg.Version != 3 {
		t.Fatalf("Version = %d", agg.Version)
	}
}

func TestAggregate_Replay_ReturnAndResubmit(t *testing.T) {
	orgID := uuid.New()
	reqID := uuid.New()
	base := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	createdPayload, _ := json.Marshal(CreatedPayload{Title: "Demo"})
	returnedPayload, _ := json.Marshal(ReturnedPayload{Reason: "needs changes"})

	agg := NewAggregate()
	err := agg.Replay([]Event{
		{OrgID: orgID, RequestID: reqID, Version: 1, Type: EventTypeCreated, OccurredAt: base.Add(1 * time.Second), Payload: createdPayload},
		{OrgID: orgID, RequestID: reqID, Version: 2, Type: EventTypeSubmitted, OccurredAt: base.Add(2 * time.Second)},
		{OrgID: orgID, RequestID: reqID, Version: 3, Type: EventTypeReturned, OccurredAt: base.Add(3 * time.Second), Payload: returnedPayload},
		{OrgID: orgID, RequestID: reqID, Version: 4, Type: EventTypeResubmitted, OccurredAt: base.Add(4 * time.Second)},
	})
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}

	if agg.Request.Status != StatusSubmitted {
		t.Fatalf("Status = %q", agg.Request.Status)
	}
	if agg.Request.SubmittedAt == nil {
		t.Fatalf("SubmittedAt is nil")
	}
	if want := base.Add(4 * time.Second); !agg.Request.SubmittedAt.Equal(want) {
		t.Fatalf("SubmittedAt = %v want %v", *agg.Request.SubmittedAt, want)
	}
	if agg.Version != 4 {
		t.Fatalf("Version = %d", agg.Version)
	}
}

func TestAggregate_RejectsInvalidTransition(t *testing.T) {
	orgID := uuid.New()
	reqID := uuid.New()
	base := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	agg := NewAggregate()
	err := agg.Replay([]Event{
		{OrgID: orgID, RequestID: reqID, Version: 1, Type: EventTypeSubmitted, OccurredAt: base.Add(1 * time.Second)},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestAggregate_RejectsVersionGap(t *testing.T) {
	orgID := uuid.New()
	reqID := uuid.New()
	base := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	createdPayload, _ := json.Marshal(CreatedPayload{Title: "Demo"})

	agg := NewAggregate()
	err := agg.Replay([]Event{
		{OrgID: orgID, RequestID: reqID, Version: 1, Type: EventTypeCreated, OccurredAt: base.Add(1 * time.Second), Payload: createdPayload},
		{OrgID: orgID, RequestID: reqID, Version: 3, Type: EventTypeSubmitted, OccurredAt: base.Add(2 * time.Second)},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}
