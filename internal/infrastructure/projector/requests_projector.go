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
				// Step 1: Submit
				_ = q.InsertRequestStepIfNotExists(ctx, dbgen.InsertRequestStepIfNotExistsParams{
					RequestID:        e.AggregateID,
					StepIndex:        1,
					Label:            "Submit",
					Status:           "pending",
					AssignedToUserID: uuid.NullUUID{},
					UpdatedAt:        e.OccurredAt,
				})

				approvalSteps := []workflowTemplateApprovalStep{{Approvers: nil}}
				if p.WorkflowTemplateID != "" {
					if tplID, err := uuid.Parse(p.WorkflowTemplateID); err == nil {
						if tpl, err := q.GetWorkflowTemplateByOrgAndID(ctx, dbgen.GetWorkflowTemplateByOrgAndIDParams{OrgID: e.OrgID, ID: tplID}); err == nil {
							if steps, err := parseWorkflowTemplateApprovalSteps(tpl.Definition); err == nil && len(steps) > 0 {
								approvalSteps = steps
							}
						}
					}
				}

				// approval steps from template (fallback: single Approval)
				for i, s := range approvalSteps {
					label := fmt.Sprintf("Approval %d", i+1)
					if len(s.Approvers) > 0 {
						label = fmt.Sprintf("Approval %d (%d approvers)", i+1, len(s.Approvers))
					}
					_ = q.InsertRequestStepIfNotExists(ctx, dbgen.InsertRequestStepIfNotExistsParams{
						RequestID:        e.AggregateID,
						StepIndex:        int32(i + 2),
						Label:            label,
						Status:           "pending",
						AssignedToUserID: uuid.NullUUID{},
						UpdatedAt:        e.OccurredAt,
					})
				}

			case request.EventTypeSubmitted:
				_, err := q.SetRequestSubmitted(ctx, dbgen.SetRequestSubmittedParams{
					OrgID:       e.OrgID,
					ID:          e.AggregateID,
					SubmittedAt: sql.NullTime{Time: e.OccurredAt, Valid: true},
				})
				if err != nil {
					return err
				}

			case request.EventTypeReturned:
				_, err := q.SetRequestReturned(ctx, dbgen.SetRequestReturnedParams{
					OrgID:     e.OrgID,
					ID:        e.AggregateID,
					UpdatedAt: e.OccurredAt,
				})
				if err != nil {
					return err
				}

			case request.EventTypeResubmitted:
				_, err := q.SetRequestResubmitted(ctx, dbgen.SetRequestResubmittedParams{
					OrgID:       e.OrgID,
					ID:          e.AggregateID,
					SubmittedAt: e.OccurredAt,
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

			case request.EventTypeRequestCommented:
				// コメント投影
				p, err := decodeJSON[request.RequestCommentedPayload](e.Payload)
				if err != nil {
					return err
				}
				commentID, err := uuid.Parse(p.CommentID)
				if err != nil {
					return err
				}
				userID, err := uuid.Parse(p.UserID)
				if err != nil {
					return err
				}
				err = q.InsertCommentIfNotExists(ctx, dbgen.InsertCommentIfNotExistsParams{
					ID:        commentID,
					OrgID:     e.OrgID,
					RequestID: e.AggregateID,
					UserID:    userID,
					Content:   p.Content,
					CreatedAt: p.CommentedAt,
				})
				if err != nil {
					return err
				}
				// コメントは監査ログに記録しない
				return nil

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

type workflowTemplateDefinition struct {
	Version int                         `json:"version"`
	Steps   []workflowTemplateStepUnion `json:"steps"`
}

type workflowTemplateStepUnion struct {
	Type      string   `json:"type"`
	Approvers []string `json:"approvers"`
}

type workflowTemplateApprovalStep struct {
	Approvers []string
}

func parseWorkflowTemplateApprovalSteps(raw json.RawMessage) ([]workflowTemplateApprovalStep, error) {
	if len(raw) == 0 {
		return nil, errors.New("missing template definition")
	}
	var d workflowTemplateDefinition
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, err
	}
	out := make([]workflowTemplateApprovalStep, 0, len(d.Steps))
	for _, s := range d.Steps {
		if s.Type != "approval" {
			continue
		}
		out = append(out, workflowTemplateApprovalStep{Approvers: s.Approvers})
	}
	return out, nil
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
