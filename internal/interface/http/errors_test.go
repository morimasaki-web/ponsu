package http

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteBadRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	msg := "bad request test"

	writeBadRequest(w, req, msg)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", res.StatusCode)
	}
	if !strings.Contains(string(body), msg) {
		t.Errorf("レスポンス本文にmsgが含まれていません: got %q", string(body))
	}
}

func TestWriteUnauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	writeUnauthorized(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", res.StatusCode)
	}
}

func TestWriteInternalError(t *testing.T) {
	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := errors.New("テスト用エラー")

	writeInternalError(w, req, logger, err)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", res.StatusCode)
	}

	if !strings.Contains(string(body), "internal server error") {
		t.Errorf("レスポンス本文にinternal server errorが含まれていません: got %q", string(body))
	}

	if strings.Contains(string(body), err.Error()) {
		t.Errorf("レスポンス本文にエラー内容が含まれています: got %q", string(body))
	}
}
