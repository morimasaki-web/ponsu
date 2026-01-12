package projector

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/morimasaki-web/ponsu/internal/domain/request"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

const RequestsProjectorName = "requests"

// NewRequestsProjector は request.* イベントを requests/read model に同期投影する。
func NewRequestsProjector() Runner {
	return Runner{
		ProjectorName: RequestsProjectorName,
		BatchSize:     200,
		Apply: func(ctx context.Context, q *dbgen.Queries, e dbgen.ListEventsForProjectorRow) error {
			actorUserID, err := metadataActorUserID(e.Metadata)
			if err != nil {
				return err
			}

			switch e.EventType {
			case request.EventTypeCreated:
				p, err := decodeJSON[request.CreatedPayload](e.Payload)
				if err != nil {
					return err
				}
				// Create request row if missing
				err = q.InsertRequestIfNotExists(ctx, dbgen.InsertRequestIfNotExistsParams{
					ID:              e.AggregateID,
					OrgID:           e.OrgID,
					Title:           p.Title,
					Status:          string(request.StatusDraft),
					CreatedByUserID: actorUserID,
					CreatedAt:       e.OccurredAt,
				})
				if err != nil {
					return err
				}
				// Minimal steps (MVP): Submit / Approve
				_ = q.InsertRequestStepIfNotExists(ctx, dbgen.InsertRequestStepIfNotExistsParams{
					RequestID:        e.AggregateID,
					StepIndex:        1,
					Label:            "Submit",
					Status:           "pending",
					AssignedToUserID: uuid.NullUUID{},
					UpdatedAt:        e.OccurredAt,
				})
				_ = q.InsertRequestStepIfNotExists(ctx, dbgen.InsertRequestStepIfNotExistsParams{
					RequestID:        e.AggregateID,
					StepIndex:        2,
					Label:            "Approve",
					Status:           "pending",
					AssignedToUserID: uuid.NullUUID{},
					UpdatedAt:        e.OccurredAt,
				})

			case request.EventTypeSubmitted:
				_, err := q.SetRequestSubmitted(ctx, dbgen.SetRequestSubmittedParams{
					OrgID:       e.OrgID,
					ID:          e.AggregateID,
					SubmittedAt: sql.NullTime{Time: e.OccurredAt, Valid: true},
				})
				if err != nil {
					return err
				}

			case request.EventTypeApproved:
				_, err := q.SetRequestApproved(ctx, dbgen.SetRequestApprovedParams{
					OrgID:     e.OrgID,
					ID:        e.AggregateID,
					DecidedAt: sql.NullTime{Time: e.OccurredAt, Valid: true},
				})
				if err != nil {
					return err
				}

			case request.EventTypeRejected:
				_, err := q.SetRequestRejected(ctx, dbgen.SetRequestRejectedParams{
					OrgID:     e.OrgID,
					ID:        e.AggregateID,
					DecidedAt: sql.NullTime{Time: e.OccurredAt, Valid: true},
				})
				if err != nil {
					return err
				}

			default:
				// ignore unrelated events
				return nil
			}

			// audit (idempotent by org+global_position)
			return q.InsertRequestAuditIfNotExists(ctx, dbgen.InsertRequestAuditIfNotExistsParams{
				OrgID:               e.OrgID,
				RequestID:           e.AggregateID,
				ActorUserID:         actorUserID,
				Action:              e.EventType,
				Data:                json.RawMessage(e.Payload),
				OccurredAt:          e.OccurredAt,
				EventGlobalPosition: sql.NullInt64{Int64: e.GlobalPosition, Valid: true},
			})
		},
	}
}

type metadataActor struct {
	ActorUserID string `json:"actor_user_id"`
}

func metadataActorUserID(raw json.RawMessage) (uuid.NullUUID, error) {
	if len(raw) == 0 {
		return uuid.NullUUID{}, nil
	}
	var m metadataActor
	if err := json.Unmarshal(raw, &m); err != nil {
		return uuid.NullUUID{}, fmt.Errorf("invalid metadata: %w", err)
	}
	if m.ActorUserID == "" {
		return uuid.NullUUID{}, nil
	}
	id, err := uuid.Parse(m.ActorUserID)
	if err != nil {
		return uuid.NullUUID{}, fmt.Errorf("invalid actor_user_id: %w", err)
	}
	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

func decodeJSON[T any](raw json.RawMessage) (T, error) {
	var out T
	if len(raw) == 0 {
		return out, errors.New("missing payload")
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}
