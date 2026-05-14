package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("boom")
}

func TestRecoverMiddleware(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler := RecoverMiddleware(logger, http.HandlerFunc(panicHandler))

	handler.ServeHTTP(w, req)

	res := w.Result()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	_ = res.Body.Close()
	body := string(bodyBytes)

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", res.StatusCode)
	}
	if !strings.Contains(body, "internal server error") {
		t.Errorf("body should contain generic message, got %q", body)
	}
	if strings.Contains(body, "boom") {
		t.Errorf("body should not contain panic details, got %q", body)
	}
}
