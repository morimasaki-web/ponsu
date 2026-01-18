package request

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransition = errors.New("invalid request state transition")
	ErrInvalidEvent      = errors.New("invalid request event")
)

// Aggregate は Request の状態をイベントから導出（replay）する。
// MVP-031 では「状態遷移が正しいこと」を最優先にする。

type Aggregate struct {
	Request Request

	// Version は replay 済みの最新 version（イベント未適用なら 0）。
	Version int32

	initialized bool
}

func NewAggregate() *Aggregate {
	return &Aggregate{}
}

// Apply はイベント1件を状態に適用する（replay / live どちらにも使う）。
func (a *Aggregate) Apply(e Event) error {
	// version monotonic check (best-effort)
	if a.Version != 0 && e.Version != a.Version+1 {
		return fmt.Errorf("%w: unexpected version: got %d want %d", ErrInvalidEvent, e.Version, a.Version+1)
	}

	switch e.Type {
	case EventTypeCreated:
		if a.initialized {
			return fmt.Errorf("%w: already created", ErrInvalidTransition)
		}
		p, err := decodePayload[CreatedPayload](e.Payload)
		if err != nil {
			return err
		}
		a.Request.OrgID = e.OrgID
		a.Request.ID = e.RequestID
		a.Request.Title = p.Title
		a.Request.Status = StatusDraft
		a.Request.CreatedAt = e.OccurredAt
		a.initialized = true

	case EventTypeSubmitted:
		if !a.initialized || a.Request.Status != StatusDraft {
			return fmt.Errorf("%w: submit requires draft", ErrInvalidTransition)
		}
		a.Request.Status = StatusSubmitted
		t := e.OccurredAt
		a.Request.SubmittedAt = &t

	case EventTypeReturned:
		if !a.initialized || a.Request.Status != StatusSubmitted {
			return fmt.Errorf("%w: return requires submitted", ErrInvalidTransition)
		}
		if _, err := decodePayload[ReturnedPayload](e.Payload); err != nil {
			return err
		}
		a.Request.Status = StatusReturned

	case EventTypeResubmitted:
		if !a.initialized || a.Request.Status != StatusReturned {
			return fmt.Errorf("%w: resubmit requires returned", ErrInvalidTransition)
		}
		a.Request.Status = StatusSubmitted
		t := e.OccurredAt
		a.Request.SubmittedAt = &t

	case EventTypeApproved:
		if !a.initialized || a.Request.Status != StatusSubmitted {
			return fmt.Errorf("%w: approve requires submitted", ErrInvalidTransition)
		}
		a.Request.Status = StatusApproved
		t := e.OccurredAt
		a.Request.DecidedAt = &t

	case EventTypeRejected:
		if !a.initialized || a.Request.Status != StatusSubmitted {
			return fmt.Errorf("%w: reject requires submitted", ErrInvalidTransition)
		}
		// payload is optional in MVP
		if _, err := decodePayload[RejectedPayload](e.Payload); err != nil {
			return err
		}
		a.Request.Status = StatusRejected
		t := e.OccurredAt
		a.Request.DecidedAt = &t

	default:
		return fmt.Errorf("%w: unknown event type: %s", ErrInvalidEvent, e.Type)
	}

	a.Version = e.Version
	return nil
}

// Replay はイベント列を順に適用する。
func (a *Aggregate) Replay(events []Event) error {
	for _, e := range events {
		if err := a.Apply(e); err != nil {
			return err
		}
	}
	return nil
}
