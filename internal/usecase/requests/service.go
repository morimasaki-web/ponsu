package requests

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/morimasaki-web/ponsu/internal/domain/request"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/eventstore"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/projector"
)

const requestAggregateType = "request"

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
)

type Service struct {
	DB            *sql.DB
	Notifier      Notifier
	PublicBaseURL string
	ActorDisplay  string
	Projector     RequestsProjector
}

type RequestsProjector interface {
	CatchUpTx(ctx context.Context, dbtx dbgen.DBTX, orgID uuid.UUID) (total int, lastPosition int64, err error)
}

func (s Service) projectorOrDefault() RequestsProjector {
	if s.Projector != nil {
		return s.Projector
	}
	r := projector.NewRequestsProjector()
	return r
}

func (s Service) CreateRequest(ctx context.Context, orgID, actorUserID uuid.UUID, title string) (uuid.UUID, error) {
	return s.CreateRequestWithTemplate(ctx, orgID, actorUserID, title, uuid.Nil)
}

func (s Service) CreateRequestWithTemplate(ctx context.Context, orgID, actorUserID uuid.UUID, title string, workflowTemplateID uuid.UUID) (uuid.UUID, error) {
	if s.DB == nil {
		return uuid.UUID{}, errors.New("db is nil")
	}
	if title == "" {
		return uuid.UUID{}, errors.New("title is required")
	}

	requestID := uuid.New()

	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback() }()

	q := dbgen.New(tx)
	m, err := q.GetMembershipByOrgAndUserID(ctx, dbgen.GetMembershipByOrgAndUserIDParams{OrgID: orgID, UserID: actorUserID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.UUID{}, ErrForbidden
		}
		return uuid.UUID{}, err
	}
	_ = m // membership existence is enough for create in MVP

	if workflowTemplateID != uuid.Nil {
		if _, err := q.GetWorkflowTemplateByOrgAndID(ctx, dbgen.GetWorkflowTemplateByOrgAndIDParams{OrgID: orgID, ID: workflowTemplateID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return uuid.UUID{}, errors.New("workflow template not found")
			}
			return uuid.UUID{}, err
		}
	}

	workflowTemplateIDStr := ""
	if workflowTemplateID != uuid.Nil {
		workflowTemplateIDStr = workflowTemplateID.String()
	}

	payload, err := json.Marshal(request.CreatedPayload{Title: title, WorkflowTemplateID: workflowTemplateIDStr})
	if err != nil {
		return uuid.UUID{}, err
	}
	metadata, err := json.Marshal(map[string]any{"actor_user_id": actorUserID.String()})
	if err != nil {
		return uuid.UUID{}, err
	}

	if _, err := eventstore.Append(ctx, tx, orgID, requestAggregateType, requestID, 0, request.EventTypeCreated, payload, metadata); err != nil {
		return uuid.UUID{}, err
	}

	// 同一TX内で直近のイベントまで投影して read model を同期させる
	if _, _, err := s.projectorOrDefault().CatchUpTx(ctx, tx, orgID); err != nil {
		return uuid.UUID{}, err
	}

	if err := tx.Commit(); err != nil {
		return uuid.UUID{}, err
	}
	return requestID, nil
}

func (s Service) SubmitRequest(ctx context.Context, orgID, actorUserID, requestID uuid.UUID) error {
	return s.applyTransition(ctx, orgID, actorUserID, requestID, request.EventTypeSubmitted, nil, false)
}

func (s Service) ApproveRequest(ctx context.Context, orgID, actorUserID, requestID uuid.UUID) error {
	return s.applyTransition(ctx, orgID, actorUserID, requestID, request.EventTypeApproved, nil, true)
}

func (s Service) RejectRequest(ctx context.Context, orgID, actorUserID, requestID uuid.UUID, reason string) error {
	payload, err := json.Marshal(request.RejectedPayload{Reason: reason})
	if err != nil {
		return err
	}
	return s.applyTransition(ctx, orgID, actorUserID, requestID, request.EventTypeRejected, payload, true)
}

func (s Service) ReturnRequest(ctx context.Context, orgID, actorUserID, requestID uuid.UUID, reason string) error {
	payload, err := json.Marshal(request.ReturnedPayload{Reason: reason})
	if err != nil {
		return err
	}
	return s.applyTransition(ctx, orgID, actorUserID, requestID, request.EventTypeReturned, payload, true)
}

func (s Service) ResubmitRequest(ctx context.Context, orgID, actorUserID, requestID uuid.UUID) error {
	if s.DB == nil {
		return errors.New("db is nil")
	}

	q := dbgen.New(s.DB)
	rm, err := q.GetRequestByOrgAndID(ctx, dbgen.GetRequestByOrgAndIDParams{OrgID: orgID, ID: requestID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if rm.CreatedByUserID.Valid && rm.CreatedByUserID.UUID != actorUserID {
		return ErrForbidden
	}

	return s.applyTransition(ctx, orgID, actorUserID, requestID, request.EventTypeResubmitted, nil, false)
}

func (s Service) maybeNotify(ctx context.Context, kind NotificationKind, orgID, actorUserID, requestID uuid.UUID) {
	if s.Notifier == nil {
		return
	}
	if s.DB == nil {
		return
	}

	q := dbgen.New(s.DB)
	req, err := q.GetRequestByOrgAndID(ctx, dbgen.GetRequestByOrgAndIDParams{OrgID: orgID, ID: requestID})
	if err != nil {
		return
	}
	actor := strings.TrimSpace(s.ActorDisplay)
	if actor == "" {
		actor = actorUserID.String()
	}

	base := strings.TrimRight(strings.TrimSpace(s.PublicBaseURL), "/")
	url := base + "/org/" + orgID.String() + "/requests/" + requestID.String()

	_ = s.Notifier.Notify(ctx, Notification{
		Kind:         kind,
		OrgID:        orgID,
		RequestID:    requestID,
		RequestTitle: req.Title,
		Actor:        actor,
		URL:          url,
	})
}

func (s Service) applyTransition(
	ctx context.Context,
	orgID uuid.UUID,
	actorUserID uuid.UUID,
	requestID uuid.UUID,
	eventType string,
	payload json.RawMessage,
	requireAdmin bool,
) error {
	if s.DB == nil {
		return errors.New("db is nil")
	}

	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	q := dbgen.New(tx)
	m, err := q.GetMembershipByOrgAndUserID(ctx, dbgen.GetMembershipByOrgAndUserIDParams{OrgID: orgID, UserID: actorUserID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}
	if requireAdmin && m.Role != "admin" {
		return ErrForbidden
	}

	rows, err := eventstore.ListByAggregate(ctx, tx, orgID, requestAggregateType, requestID, 1)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return ErrNotFound
	}
	for _, r := range rows {
		if r.AggregateType != requestAggregateType {
			return fmt.Errorf("unexpected aggregate_type: %s", r.AggregateType)
		}
	}

	domainEvents, err := request.FromEventStoreRows(rows)
	if err != nil {
		return err
	}
	agg := request.NewAggregate()
	if err := agg.Replay(domainEvents); err != nil {
		return err
	}

	expectedVersion := agg.Version
	if payload == nil {
		payload = json.RawMessage("{}")
	}

	// 先にドメインで遷移を検証（payload decode も含む）
	nextVersion := expectedVersion + 1
	if err := agg.Apply(request.Event{OrgID: orgID, RequestID: requestID, Version: nextVersion, Type: eventType, OccurredAt: time.Now(), Payload: payload}); err != nil {
		return err
	}

	metadata, err := json.Marshal(map[string]any{"actor_user_id": actorUserID.String()})
	if err != nil {
		return err
	}

	// append（楽観ロック expectedVersion=agg.Version）
	if _, err := eventstore.Append(ctx, tx, orgID, requestAggregateType, requestID, expectedVersion, eventType, payload, metadata); err != nil {
		return err
	}

	if _, _, err := s.projectorOrDefault().CatchUpTx(ctx, tx, orgID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	switch eventType {
	case request.EventTypeSubmitted:
		s.maybeNotify(ctx, NotificationKindSubmitted, orgID, actorUserID, requestID)
	case request.EventTypeApproved:
		s.maybeNotify(ctx, NotificationKindApproved, orgID, actorUserID, requestID)
	case request.EventTypeRejected:
		s.maybeNotify(ctx, NotificationKindRejected, orgID, actorUserID, requestID)
	}
	return nil
}
