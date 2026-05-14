package projector

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

func TestRunner_RunOnceTx_NoEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	r := Runner{ProjectorName: "test", BatchSize: 100, Apply: func(ctx context.Context, q *dbgen.Queries, e dbgen.ListEventsForProjectorRow) error {
		t.Fatalf("Apply should not be called")
		return nil
	}}

	// GetProjectionCheckpoint -> no rows
	mock.ExpectQuery("(?s)FROM public\\.projection_checkpoints").
		WithArgs(orgID, r.ProjectorName).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "projector_name", "last_position", "updated_at"}))

	// ListEventsForProjector -> empty
	mock.ExpectQuery("(?s)FROM public\\.event_store").
		WithArgs(orgID, int64(0), int32(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "global_position", "event_type", "payload", "metadata", "occurred_at"}))

	processed, lastPosition, err := r.RunOnceTx(context.Background(), db, orgID)
	if err != nil {
		t.Fatalf("RunOnceTx() error = %v", err)
	}
	if processed != 0 {
		t.Fatalf("processed = %d", processed)
	}
	if lastPosition != 0 {
		t.Fatalf("lastPosition = %d", lastPosition)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestRunner_RunOnceTx_ApplyAndCheckpoint(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	occurredAt := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)
	eventID := uuid.New()
	aggID := uuid.New()

	applyCalled := 0
	r := Runner{ProjectorName: "test", BatchSize: 100, Apply: func(ctx context.Context, q *dbgen.Queries, e dbgen.ListEventsForProjectorRow) error {
		applyCalled++
		return nil
	}}

	// GetProjectionCheckpoint -> no rows
	mock.ExpectQuery("(?s)FROM public\\.projection_checkpoints").
		WithArgs(orgID, r.ProjectorName).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "projector_name", "last_position", "updated_at"}))

	// ListEventsForProjector -> 1 row
	mock.ExpectQuery("(?s)FROM public\\.event_store").
		WithArgs(orgID, int64(0), int32(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "global_position", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(eventID, orgID, "request", aggID, int32(1), int64(10), "request.created", []byte("{}"), []byte("{}"), occurredAt))

	// UpsertProjectionCheckpoint -> 1 row
	mock.ExpectQuery("(?s)INSERT INTO public\\.projection_checkpoints.*RETURNING org_id").
		WithArgs(orgID, r.ProjectorName, int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "projector_name", "last_position", "updated_at"}).
			AddRow(orgID, "test", int64(10), occurredAt))

	processed, lastPosition, err := r.RunOnceTx(context.Background(), db, orgID)
	if err != nil {
		t.Fatalf("RunOnceTx() error = %v", err)
	}
	if processed != 1 {
		t.Fatalf("processed = %d", processed)
	}
	if lastPosition != 10 {
		t.Fatalf("lastPosition = %d", lastPosition)
	}
	if applyCalled != 1 {
		t.Fatalf("applyCalled = %d", applyCalled)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}
