package http

import (
	"log/slog"
	"net/http"
)

func writeBadRequest(w http.ResponseWriter, r *http.Request, msg string) {
	_ = r
	http.Error(w, msg, http.StatusBadRequest)
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	_ = r
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

func writeForbidden(w http.ResponseWriter, r *http.Request) {
	_ = r
	http.Error(w, "forbidden", http.StatusForbidden)
}

func writeNotFound(w http.ResponseWriter, r *http.Request) {
	_ = r
	http.Error(w, "not found", http.StatusNotFound)
}

func writeInternalError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	if logger == nil {
		logger = slog.Default()
	}
	if err != nil {
		logger.Error(
			"http handler error",
			"method", r.Method,
			"path", r.URL.Path,
			"status", http.StatusInternalServerError,
			"error", err,
		)
	}
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
