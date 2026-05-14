package requests

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/morimasaki-web/ponsu/internal/domain/request"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/eventstore"
)

type noopProjector struct {
	calls int
}

func (p *noopProjector) CatchUpTx(ctx context.Context, dbtx dbgen.DBTX, orgID uuid.UUID) (int, int64, error) {
	p.calls++
	return 0, 0, nil
}

func TestService_SubmitRequest_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	actorUserID := uuid.New()
	requestID := uuid.New()

	createdPayload, _ := json.Marshal(request.CreatedPayload{Title: "Demo"})
	base := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)

	p := &noopProjector{}
	s := Service{DB: db, Projector: p}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM public\\.memberships").
		WithArgs(orgID, actorUserID).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "role"}).AddRow(orgID, actorUserID, "member"))

	mock.ExpectQuery("FROM public\\.event_store").
		WithArgs(orgID, "request", requestID, int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(1), request.EventTypeCreated, createdPayload, []byte("{}"), base))

	mock.ExpectQuery("(?s)WITH current AS .*INSERT INTO public\\.event_store.*RETURNING").
		WithArgs(orgID, "request", requestID, int32(2), request.EventTypeSubmitted, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(2), request.EventTypeSubmitted, []byte("{}"), []byte("{}"), base.Add(1*time.Second)))

	mock.ExpectCommit()

	if err := s.SubmitRequest(context.Background(), orgID, actorUserID, requestID); err != nil {
		t.Fatalf("SubmitRequest() error = %v", err)
	}
	if p.calls != 1 {
		t.Fatalf("projector calls = %d", p.calls)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestService_ApproveRequest_RequiresAdmin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	actorUserID := uuid.New()
	requestID := uuid.New()

	p := &noopProjector{}
	s := Service{DB: db, Projector: p}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM public\\.memberships").
		WithArgs(orgID, actorUserID).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "role"}).AddRow(orgID, actorUserID, "member"))
	mock.ExpectRollback()

	err = s.ApproveRequest(context.Background(), orgID, actorUserID, requestID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
	if p.calls != 0 {
		t.Fatalf("projector calls = %d", p.calls)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestService_ApproveRequest_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	actorUserID := uuid.New()
	requestID := uuid.New()

	createdPayload, _ := json.Marshal(request.CreatedPayload{Title: "Demo"})
	base := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)

	p := &noopProjector{}
	s := Service{DB: db, Projector: p}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM public\\.memberships").
		WithArgs(orgID, actorUserID).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "role"}).AddRow(orgID, actorUserID, "admin"))

	mock.ExpectQuery("FROM public\\.event_store").
		WithArgs(orgID, "request", requestID, int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(1), request.EventTypeCreated, createdPayload, []byte("{}"), base).
			AddRow(uuid.New(), orgID, "request", requestID, int32(2), request.EventTypeSubmitted, []byte("{}"), []byte("{}"), base.Add(1*time.Second)))

	mock.ExpectQuery("(?s)WITH current AS .*INSERT INTO public\\.event_store.*RETURNING").
		WithArgs(orgID, "request", requestID, int32(3), request.EventTypeApproved, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(3), request.EventTypeApproved, []byte("{}"), []byte("{}"), base.Add(2*time.Second)))

	mock.ExpectCommit()

	if err := s.ApproveRequest(context.Background(), orgID, actorUserID, requestID); err != nil {
		t.Fatalf("ApproveRequest() error = %v", err)
	}
	if p.calls != 1 {
		t.Fatalf("projector calls = %d", p.calls)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestService_RejectRequest_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	actorUserID := uuid.New()
	requestID := uuid.New()

	createdPayload, _ := json.Marshal(request.CreatedPayload{Title: "Demo"})
	rejectedPayload, _ := json.Marshal(request.RejectedPayload{Reason: "no"})
	base := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)

	p := &noopProjector{}
	s := Service{DB: db, Projector: p}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM public\\.memberships").
		WithArgs(orgID, actorUserID).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "role"}).AddRow(orgID, actorUserID, "admin"))

	mock.ExpectQuery("FROM public\\.event_store").
		WithArgs(orgID, "request", requestID, int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(1), request.EventTypeCreated, createdPayload, []byte("{}"), base).
			AddRow(uuid.New(), orgID, "request", requestID, int32(2), request.EventTypeSubmitted, []byte("{}"), []byte("{}"), base.Add(1*time.Second)))

	mock.ExpectQuery("(?s)WITH current AS .*INSERT INTO public\\.event_store.*RETURNING").
		WithArgs(orgID, "request", requestID, int32(3), request.EventTypeRejected, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(3), request.EventTypeRejected, rejectedPayload, []byte("{}"), base.Add(2*time.Second)))

	mock.ExpectCommit()

	if err := s.RejectRequest(context.Background(), orgID, actorUserID, requestID, "no"); err != nil {
		t.Fatalf("RejectRequest() error = %v", err)
	}
	if p.calls != 1 {
		t.Fatalf("projector calls = %d", p.calls)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestService_ReturnAndResubmit_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	actorUserID := uuid.New()
	requestID := uuid.New()

	createdPayload, _ := json.Marshal(request.CreatedPayload{Title: "Demo"})
	returnedPayload, _ := json.Marshal(request.ReturnedPayload{Reason: "needs changes"})
	base := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)

	p := &noopProjector{}
	s := Service{DB: db, Projector: p}

	// Return (admin)
	mock.ExpectBegin()
	mock.ExpectQuery("FROM public\\.memberships").
		WithArgs(orgID, actorUserID).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "role"}).AddRow(orgID, actorUserID, "admin"))

	mock.ExpectQuery("FROM public\\.event_store").
		WithArgs(orgID, "request", requestID, int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(1), request.EventTypeCreated, createdPayload, []byte("{}"), base).
			AddRow(uuid.New(), orgID, "request", requestID, int32(2), request.EventTypeSubmitted, []byte("{}"), []byte("{}"), base.Add(1*time.Second)))

	mock.ExpectQuery("(?s)WITH current AS .*INSERT INTO public\\.event_store.*RETURNING").
		WithArgs(orgID, "request", requestID, int32(3), request.EventTypeReturned, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(3), request.EventTypeReturned, returnedPayload, []byte("{}"), base.Add(2*time.Second)))

	mock.ExpectCommit()

	if err := s.ReturnRequest(context.Background(), orgID, actorUserID, requestID, "needs changes"); err != nil {
		t.Fatalf("ReturnRequest() error = %v", err)
	}

	// Resubmit (created_by must match)
	mock.ExpectQuery("FROM public\\.requests").
		WithArgs(orgID, requestID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "title", "status", "created_by_user_id", "decided_by_user_id", "created_at", "updated_at", "submitted_at", "decided_at"}).
			AddRow(requestID, orgID, "Demo", "returned", actorUserID.String(), nil, base, base, nil, nil))

	mock.ExpectBegin()
	mock.ExpectQuery("FROM public\\.memberships").
		WithArgs(orgID, actorUserID).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "role"}).AddRow(orgID, actorUserID, "member"))

	mock.ExpectQuery("FROM public\\.event_store").
		WithArgs(orgID, "request", requestID, int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(1), request.EventTypeCreated, createdPayload, []byte("{}"), base).
			AddRow(uuid.New(), orgID, "request", requestID, int32(2), request.EventTypeSubmitted, []byte("{}"), []byte("{}"), base.Add(1*time.Second)).
			AddRow(uuid.New(), orgID, "request", requestID, int32(3), request.EventTypeReturned, returnedPayload, []byte("{}"), base.Add(2*time.Second)))

	mock.ExpectQuery("(?s)WITH current AS .*INSERT INTO public\\.event_store.*RETURNING").
		WithArgs(orgID, "request", requestID, int32(4), request.EventTypeResubmitted, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(4), request.EventTypeResubmitted, []byte("{}"), []byte("{}"), base.Add(3*time.Second)))

	mock.ExpectCommit()

	if err := s.ResubmitRequest(context.Background(), orgID, actorUserID, requestID); err != nil {
		t.Fatalf("ResubmitRequest() error = %v", err)
	}

	if p.calls != 2 {
		t.Fatalf("projector calls = %d", p.calls)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestService_SubmitRequest_ReturnsVersionConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	actorUserID := uuid.New()
	requestID := uuid.New()

	createdPayload, _ := json.Marshal(request.CreatedPayload{Title: "Demo"})
	base := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)

	s := Service{DB: db, Projector: &noopProjector{}}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM public\\.memberships").
		WithArgs(orgID, actorUserID).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "role"}).AddRow(orgID, actorUserID, "member"))

	mock.ExpectQuery("FROM public\\.event_store").
		WithArgs(orgID, "request", requestID, int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(uuid.New(), orgID, "request", requestID, int32(1), request.EventTypeCreated, createdPayload, []byte("{}"), base))

	// 0行 -> sql.ErrNoRows -> ErrVersionConflict
	mock.ExpectQuery("(?s)WITH current AS .*INSERT INTO public\\.event_store.*RETURNING").
		WithArgs(orgID, "request", requestID, int32(2), request.EventTypeSubmitted, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}))

	mock.ExpectRollback()

	err = s.SubmitRequest(context.Background(), orgID, actorUserID, requestID)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, eventstore.ErrVersionConflict) {
		t.Fatalf("error = %v, want eventstore.ErrVersionConflict", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}
