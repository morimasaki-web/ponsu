package http

import (
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
)

func TestBuildInfo(t *testing.T) {

	cfg := config.Config{}
	logger := slog.Default()
	var db *sql.DB

	mux := NewMux(cfg, logger, db)

	req := httptest.NewRequest("GET", "/buildinfo", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("content-type"); !strings.Contains(ct, "application/json") {
		t.Errorf("content-type = %s, want application/json", ct)
	}
}
