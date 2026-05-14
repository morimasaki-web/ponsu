package http

import (
	"database/sql"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
)

func TestRequestIdMiddleware(t *testing.T) {
	cfg := config.Config{}
	logger := slog.Default()
	var db *sql.DB
	mux := NewMux(cfg, logger, db)

	t.Run("request headerあり", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/buildinfo", nil)
		expectedHeader := "test-request-id"
		req.Header.Set("X-Request-ID", expectedHeader)
		handler := RequestIDMiddleware(mux)
		handler.ServeHTTP(w, req)
		res := w.Result()
		if res.Header.Get("X-Request-ID") != expectedHeader {
			t.Errorf("リクエストと同じヘッダがレスポンスに含まれていません。got=%s, want=%s", res.Header.Get("X-Request-ID"), expectedHeader)
		}
	})

	t.Run("request headerなし", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/buildinfo", nil)
		handler := RequestIDMiddleware(mux)
		handler.ServeHTTP(w, req)
		res := w.Result()
		got := res.Header.Get("X-Request-ID")
		if got == "" {
			t.Errorf("レスポンスヘッダにX-Request-IDが含まれていません。")
		}
	})
}
