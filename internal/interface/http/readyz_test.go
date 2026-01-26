package http

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
)

func TestReadyz(t *testing.T) {

	cfg := config.Config{}
	logger := slog.Default()
	var db *sql.DB

	mux := NewMux(cfg, logger, db)

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("content-type = %s, want application/json", ct)
	}

	var body struct {
		Ok bool `json:"ok"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if body.Ok != true {
		t.Errorf("ok = %v, want true", body.Ok)
	}
}
