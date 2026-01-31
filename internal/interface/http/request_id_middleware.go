package http

import (
	"crypto/rand"
	"fmt"
	"net/http"
)

const requestIDHeader = "X-Request-ID"

func generateRequestID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)

	if err != nil {
		return "unknown"
	}

	return fmt.Sprintf("%x", b)
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}
		r.Header.Set(requestIDHeader, requestID)
		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, r)
	})
}
