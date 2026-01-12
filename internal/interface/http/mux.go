// Package http は PonSu のHTTPルーティングと最小のハンドラ群を提供する。
// MVPでは healthz と OIDC 認証関連のエンドポイントを定義する。
package http

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
)

// NewMux はアプリケーションのHTTPルートを登録した ServeMux を返す。
func NewMux(cfg config.Config, logger *slog.Logger, db *sql.DB) *http.ServeMux {
	if logger == nil {
		logger = slog.Default()
	}

	auth := NewOIDCAuth(cfg, logger, db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /", auth.HandleHome)
	mux.HandleFunc("GET /auth/login", auth.HandleLogin)
	mux.HandleFunc("GET /auth/callback", auth.HandleCallback)
	mux.HandleFunc("GET /auth/logout", auth.HandleLogout)

	return mux
}
