package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestAppend_ReturnsVersionConflictOnNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	aggType := "request"
	aggID := uuid.New()
	eventType := "request.submitted"
	payload := json.RawMessage(`{"k":"v"}`)
	metadata := json.RawMessage(`{"actor_user_id":"` + uuid.New().String() + `"}`)

	// AppendEvent は QueryRowContext で実行され、結果0件の場合 Scan() が sql.ErrNoRows になる。
	mock.ExpectQuery("(?s)WITH current AS .*INSERT INTO public\\.event_store.*RETURNING id").
		WithArgs(orgID, aggType, aggID, int32(1), eventType, payload, metadata).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}))

	_, err = Append(context.Background(), db, orgID, aggType, aggID, 0, eventType, payload, metadata)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("error = %v, want ErrVersionConflict", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestAppend_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	aggType := "request"
	aggID := uuid.New()
	eventType := "request.created"
	payload := json.RawMessage(`{"title":"Demo"}`)
	metadata := json.RawMessage(`{"actor_user_id":"` + uuid.New().String() + `"}`)

	id := uuid.New()
	occurredAt := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("(?s)WITH current AS .*INSERT INTO public\\.event_store.*RETURNING id").
		WithArgs(orgID, aggType, aggID, int32(1), eventType, payload, metadata).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at"}).
			AddRow(id, orgID, aggType, aggID, int32(1), eventType, payload, metadata, occurredAt))

	row, err := Append(context.Background(), db, orgID, aggType, aggID, 0, eventType, payload, metadata)
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if row.ID != id {
		t.Fatalf("ID = %v", row.ID)
	}
	if row.Version != 1 {
		t.Fatalf("Version = %d", row.Version)
	}
	if row.EventType != eventType {
		t.Fatalf("EventType = %q", row.EventType)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestListByAggregateID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	aggID := uuid.New()

	eventRows := sqlmock.NewRows([]string{
		"id", "org_id", "aggregate_type", "aggregate_id", "version", "event_type", "payload", "metadata", "occurred_at",
	}).
		AddRow(uuid.New(), uuid.New(), "request", aggID, 1, "request.created", json.RawMessage(`{"title":"A"}`), json.RawMessage(`{}`), time.Now()).
		AddRow(uuid.New(), uuid.New(), "request", aggID, 2, "request.updated", json.RawMessage(`{"title":"B"}`), json.RawMessage(`{}`), time.Now())

	mock.ExpectQuery("SELECT (.+) FROM public.event_store WHERE aggregate_id = (.+) ORDER BY (.+)").
		WithArgs(aggID).
		WillReturnRows(eventRows)

	rows, err := ListByAggregateID(context.Background(), db, aggID)
	if err != nil {
		t.Fatalf("ListByAggregateID() error = %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}
