package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
)

func TestHandleRequestsCreate_IdempotencyKeyConflictOnDifferentBody(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	h := NewRequestsHandler(config.Config{}, db, nil)

	orgID := uuid.New()
	userID := uuid.New()
	templateID := uuid.New()

	form := url.Values{}
	form.Set("title", "Request A")
	form.Set("workflow_template_id", templateID.String())

	req := httptest.NewRequest(http.MethodPost, "/org/"+orgID.String()+"/requests", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", "key-1")
	req.SetPathValue("orgID", orgID.String())

	w := httptest.NewRecorder()
	sess := sessionData{OrgID: orgID.String(), UserID: userID.String()}

	mock.ExpectExec("INSERT INTO idempotency").
		WithArgs(orgID, userID, "create_request", "key-1", sha256Hash("Request A\n"+templateID.String()), int32(0), "").
		WillReturnResult(sqlmock.NewResult(0, 0))

	rows := sqlmock.NewRows([]string{
		"org_id",
		"actor_user_id",
		"action",
		"idempotency_key",
		"request_hash",
		"status_code",
		"response_body",
		"created_at",
	}).
		AddRow(orgID, userID, "create_request", "key-1", "different-hash", int32(http.StatusSeeOther), "/org/"+orgID.String()+"/requests/old", time.Now())

	mock.ExpectQuery("FROM idempotency").
		WithArgs(orgID, userID, "create_request", "key-1").
		WillReturnRows(rows)

	h.HandleRequestsCreate(w, req, sess)

	res := w.Result()
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusConflict)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestHandleRequestsCreate_IdempotencyKeyReturnsSavedRedirect(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	h := NewRequestsHandler(config.Config{}, db, nil)

	orgID := uuid.New()
	userID := uuid.New()
	templateID := uuid.New()
	requestHash := sha256Hash("Request A\n" + templateID.String())
	savedRedirect := "/org/" + orgID.String() + "/requests/" + uuid.NewString()

	form := url.Values{}
	form.Set("title", "Request A")
	form.Set("workflow_template_id", templateID.String())

	req := httptest.NewRequest(http.MethodPost, "/org/"+orgID.String()+"/requests", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", "key-1")
	req.SetPathValue("orgID", orgID.String())

	w := httptest.NewRecorder()
	sess := sessionData{OrgID: orgID.String(), UserID: userID.String()}

	mock.ExpectExec("INSERT INTO idempotency").
		WithArgs(orgID, userID, "create_request", "key-1", requestHash, int32(0), "").
		WillReturnResult(sqlmock.NewResult(0, 0))

	rows := sqlmock.NewRows([]string{
		"org_id",
		"actor_user_id",
		"action",
		"idempotency_key",
		"request_hash",
		"status_code",
		"response_body",
		"created_at",
	}).
		AddRow(orgID, userID, "create_request", "key-1", requestHash, int32(http.StatusSeeOther), savedRedirect, time.Now())

	mock.ExpectQuery("FROM idempotency").
		WithArgs(orgID, userID, "create_request", "key-1").
		WillReturnRows(rows)

	h.HandleRequestsCreate(w, req, sess)

	res := w.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusSeeOther)
	}
	if got := res.Header.Get("Location"); got != savedRedirect {
		t.Fatalf("location = %q, want %q", got, savedRedirect)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestHandleRequestsCreate_WithoutIdempotencyKey_ProceedsNormalFlow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	h := NewRequestsHandler(config.Config{}, db, nil)

	orgID := uuid.New()
	userID := uuid.New()
	templateID := uuid.New()

	form := url.Values{}
	form.Set("title", "Request A")
	form.Set("workflow_template_id", templateID.String())

	req := httptest.NewRequest(http.MethodPost, "/org/"+orgID.String()+"/requests", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("orgID", orgID.String())

	w := httptest.NewRecorder()
	sess := sessionData{OrgID: orgID.String(), UserID: userID.String()}

	// service.CreateRequestWithTemplate -> BeginTx -> membership check (no rows) -> rollback
	mock.ExpectBegin()
	mock.ExpectQuery("FROM public\\.memberships").
		WithArgs(orgID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "role"}))
	mock.ExpectRollback()

	// handler error path renders new page and reloads templates
	templateRows := sqlmock.NewRows([]string{
		"id",
		"org_id",
		"name",
		"description",
		"definition",
		"created_by_user_id",
		"created_at",
		"updated_at",
	})
	mock.ExpectQuery("FROM public\\.workflow_templates").
		WithArgs(orgID, int32(200), int32(0)).
		WillReturnRows(templateRows)

	h.HandleRequestsCreate(w, req, sess)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}
