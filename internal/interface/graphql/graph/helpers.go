package graph

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
	"github.com/morimasaki-web/ponsu/internal/interface/graphql/graph/model"
	"github.com/morimasaki-web/ponsu/internal/interface/graphqlctx"
)

var (
	errUnauthorized = errors.New("unauthorized")
	errForbidden    = errors.New("forbidden")
)

func viewerFromContext(ctx context.Context) (graphqlctx.Viewer, error) {
	v, ok := graphqlctx.ViewerFrom(ctx)
	if !ok {
		return graphqlctx.Viewer{}, errUnauthorized
	}
	if v.UserID == uuid.Nil || v.OrgID == uuid.Nil {
		return graphqlctx.Viewer{}, errUnauthorized
	}
	return v, nil
}

func actorDisplayFromViewer(v graphqlctx.Viewer) string {
	name := strings.TrimSpace(v.Name)
	email := strings.TrimSpace(v.Email)
	if name != "" && email != "" {
		return name + " <" + email + ">"
	}
	if email != "" {
		return email
	}
	if name != "" {
		return name
	}
	return ""
}

func (r *Resolver) queries() (*dbgen.Queries, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("db is not configured")
	}
	return dbgen.New(r.DB), nil
}

func parseUUID(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, errors.New("invalid id")
	}
	return uid, nil
}

func mapNullUUID(u uuid.NullUUID) *string {
	if !u.Valid {
		return nil
	}
	s := u.UUID.String()
	return &s
}

func mapRequestRow(row dbgen.Request) *model.Request {
	out := &model.Request{
		ID:              row.ID.String(),
		OrgID:           row.OrgID.String(),
		Title:           row.Title,
		Status:          row.Status,
		CreatedByUserID: mapNullUUID(row.CreatedByUserID),
		DecidedByUserID: mapNullUUID(row.DecidedByUserID),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		Steps:           []*model.RequestStep{},
		AuditTrail:      []*model.RequestAudit{},
	}
	if row.SubmittedAt.Valid {
		t := row.SubmittedAt.Time
		out.SubmittedAt = &t
	}
	if row.DecidedAt.Valid {
		t := row.DecidedAt.Time
		out.DecidedAt = &t
	}
	return out
}

func mapRequestSteps(rows []dbgen.RequestStep) []*model.RequestStep {
	out := make([]*model.RequestStep, 0, len(rows))
	for _, r := range rows {
		step := &model.RequestStep{
			StepIndex: int(r.StepIndex),
			Label:     r.Label,
			Status:    r.Status,
			UpdatedAt: r.UpdatedAt,
		}
		if r.AssignedToUserID.Valid {
			s := r.AssignedToUserID.UUID.String()
			step.AssignedToUserID = &s
		}
		out = append(out, step)
	}
	return out
}

func mapAuditTrail(rows []dbgen.ListRequestAuditTrailRow) ([]*model.RequestAudit, error) {
	out := make([]*model.RequestAudit, 0, len(rows))
	for _, r := range rows {
		var data map[string]any
		if len(r.Data) == 0 {
			data = map[string]any{}
		} else {
			if err := json.Unmarshal(r.Data, &data); err != nil {
				return nil, errors.New("invalid audit data")
			}
		}

		a := &model.RequestAudit{
			ID:         r.ID.String(),
			Action:     r.Action,
			Data:       data,
			OccurredAt: r.OccurredAt,
		}
		if r.ActorUserID.Valid {
			s := r.ActorUserID.UUID.String()
			a.ActorUserID = &s
		}
		out = append(out, a)
	}
	return out, nil
}

func mapWorkflowTemplateRow(row dbgen.WorkflowTemplate) (*model.WorkflowTemplate, error) {
	var def map[string]any
	if len(row.Definition) == 0 {
		def = map[string]any{}
	} else {
		if err := json.Unmarshal(row.Definition, &def); err != nil {
			return nil, errors.New("invalid template definition")
		}
	}

	out := &model.WorkflowTemplate{
		ID:          row.ID.String(),
		OrgID:       row.OrgID.String(),
		Name:        row.Name,
		Description: row.Description,
		Definition:  def,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	out.CreatedByUserID = mapNullUUID(row.CreatedByUserID)
	return out, nil
}
