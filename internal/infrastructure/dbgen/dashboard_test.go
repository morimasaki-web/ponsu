package dbgen

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestCountRequestsByStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	q := New(db)

	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) AS count FROM public\\.requests").
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
			AddRow("draft", int64(5)).
			AddRow("submitted", int64(10)).
			AddRow("approved", int64(15)).
			AddRow("rejected", int64(3)))

	rows, err := q.CountRequestsByStatus(context.Background(), orgID)
	if err != nil {
		t.Fatalf("CountRequestsByStatus() error = %v", err)
	}

	if len(rows) != 4 {
		t.Fatalf("len(rows) = %d, want 4", len(rows))
	}

	if rows[0].Status != "draft" || rows[0].Count != 5 {
		t.Errorf("rows[0] = %+v, want {draft 5}", rows[0])
	}
	if rows[1].Status != "submitted" || rows[1].Count != 10 {
		t.Errorf("rows[1] = %+v, want {submitted 10}", rows[1])
	}
	if rows[2].Status != "approved" || rows[2].Count != 15 {
		t.Errorf("rows[2] = %+v, want {approved 15}", rows[2])
	}
	if rows[3].Status != "rejected" || rows[3].Count != 3 {
		t.Errorf("rows[3] = %+v, want {rejected 3}", rows[3])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestCountRequestsByMonth(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	q := New(db)

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

	month1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	month2 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	month3 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT\\s+DATE_TRUNC\\('month', created_at\\)::timestamptz AS month").
		WithArgs(orgID, startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"month", "count"}).
			AddRow(month3, int64(8)).
			AddRow(month2, int64(12)).
			AddRow(month1, int64(10)))

	params := CountRequestsByMonthParams{
		OrgID:     orgID,
		StartDate: sql.NullTime{Time: startDate, Valid: true},
		EndDate:   sql.NullTime{Time: endDate, Valid: true},
	}

	rows, err := q.CountRequestsByMonth(context.Background(), params)
	if err != nil {
		t.Fatalf("CountRequestsByMonth() error = %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3", len(rows))
	}

	// ORDER BY month DESC なので降順
	if !rows[0].Month.Equal(month3) || rows[0].Count != 8 {
		t.Errorf("rows[0] = %+v, want {2026-03-01 8}", rows[0])
	}
	if !rows[1].Month.Equal(month2) || rows[1].Count != 12 {
		t.Errorf("rows[1] = %+v, want {2026-02-01 12}", rows[1])
	}
	if !rows[2].Month.Equal(month1) || rows[2].Count != 10 {
		t.Errorf("rows[2] = %+v, want {2026-01-01 10}", rows[2])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestCountRequestsByMonth_NoDateFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	q := New(db)

	month1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT\\s+DATE_TRUNC\\('month', created_at\\)::timestamptz AS month").
		WithArgs(orgID, nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{"month", "count"}).
			AddRow(month1, int64(10)))

	params := CountRequestsByMonthParams{
		OrgID: orgID,
	}

	rows, err := q.CountRequestsByMonth(context.Background(), params)
	if err != nil {
		t.Fatalf("CountRequestsByMonth() error = %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestAvgTimeToApproval(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	q := New(db)

	// 平均承認時間: 3600秒（1時間）、サンプル数: 25件
	mock.ExpectQuery("SELECT\\s+COALESCE\\(AVG\\(EXTRACT\\(EPOCH FROM").
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{"avg_seconds", "sample_count"}).
			AddRow(float64(3600.5), int64(25)))

	row, err := q.AvgTimeToApproval(context.Background(), orgID)
	if err != nil {
		t.Fatalf("AvgTimeToApproval() error = %v", err)
	}

	if row.AvgSeconds != 3600.5 {
		t.Errorf("AvgSeconds = %f, want 3600.5", row.AvgSeconds)
	}
	if row.SampleCount != 25 {
		t.Errorf("SampleCount = %d, want 25", row.SampleCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestAvgTimeToApproval_NoApprovedRequests(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	q := New(db)

	// 承認済み申請が0件の場合、COALESCE で 0 を返す
	mock.ExpectQuery("SELECT\\s+COALESCE\\(AVG\\(EXTRACT\\(EPOCH FROM").
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{"avg_seconds", "sample_count"}).
			AddRow(float64(0), int64(0)))

	row, err := q.AvgTimeToApproval(context.Background(), orgID)
	if err != nil {
		t.Fatalf("AvgTimeToApproval() error = %v", err)
	}

	if row.AvgSeconds != 0 {
		t.Errorf("AvgSeconds = %f, want 0", row.AvgSeconds)
	}
	if row.SampleCount != 0 {
		t.Errorf("SampleCount = %d, want 0", row.SampleCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestGetDashboardSummary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	q := New(db)

	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\) FILTER \\(WHERE status = 'draft'\\)::bigint AS draft_count").
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{
			"draft_count",
			"submitted_count",
			"approved_count",
			"rejected_count",
			"total_count",
			"avg_approval_seconds",
		}).AddRow(int64(5), int64(10), int64(15), int64(3), int64(33), float64(7200.25)))

	row, err := q.GetDashboardSummary(context.Background(), orgID)
	if err != nil {
		t.Fatalf("GetDashboardSummary() error = %v", err)
	}

	if row.DraftCount != 5 {
		t.Errorf("DraftCount = %d, want 5", row.DraftCount)
	}
	if row.SubmittedCount != 10 {
		t.Errorf("SubmittedCount = %d, want 10", row.SubmittedCount)
	}
	if row.ApprovedCount != 15 {
		t.Errorf("ApprovedCount = %d, want 15", row.ApprovedCount)
	}
	if row.RejectedCount != 3 {
		t.Errorf("RejectedCount = %d, want 3", row.RejectedCount)
	}
	if row.TotalCount != 33 {
		t.Errorf("TotalCount = %d, want 33", row.TotalCount)
	}
	if row.AvgApprovalSeconds != 7200.25 {
		t.Errorf("AvgApprovalSeconds = %f, want 7200.25", row.AvgApprovalSeconds)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestGetDashboardSummary_NoData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	q := New(db)

	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\) FILTER \\(WHERE status = 'draft'\\)::bigint AS draft_count").
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{
			"draft_count",
			"submitted_count",
			"approved_count",
			"rejected_count",
			"total_count",
			"avg_approval_seconds",
		}).AddRow(int64(0), int64(0), int64(0), int64(0), int64(0), float64(0)))

	row, err := q.GetDashboardSummary(context.Background(), orgID)
	if err != nil {
		t.Fatalf("GetDashboardSummary() error = %v", err)
	}

	if row.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", row.TotalCount)
	}
	if row.AvgApprovalSeconds != 0 {
		t.Errorf("AvgApprovalSeconds = %f, want 0", row.AvgApprovalSeconds)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}
