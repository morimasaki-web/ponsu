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

	// MVP-041: Request create (SSR minimal)
	mux.HandleFunc("GET /requests", auth.RequireLogin(auth.HandleRequestsIndexShortcut))
	mux.HandleFunc("GET /requests/new", auth.RequireLogin(auth.HandleRequestsNewShortcut))
	mux.HandleFunc("GET /org/{orgID}/requests", auth.RequireLogin(auth.HandleRequestsIndex))
	mux.HandleFunc("GET /org/{orgID}/requests/new", auth.RequireLogin(auth.HandleRequestsNew))
	mux.HandleFunc("POST /org/{orgID}/requests", auth.RequireLogin(auth.HandleRequestsCreate))
	mux.HandleFunc("GET /org/{orgID}/requests/{requestID}", auth.RequireLogin(auth.HandleRequestsShow))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/submit", auth.RequireLogin(auth.HandleRequestsSubmit))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/approve", auth.RequireLogin(auth.HandleRequestsApprove))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/reject", auth.RequireLogin(auth.HandleRequestsReject))

	// MVP-011: RBAC
	mux.HandleFunc("GET /admin/templates/new", auth.RequireRole("admin", auth.HandleAdminTemplatesNewShortcut))
	mux.HandleFunc("GET /org/{orgID}/admin/templates", auth.RequireRole("admin", auth.HandleAdminTemplatesIndex))
	mux.HandleFunc("GET /org/{orgID}/admin/templates/new", auth.RequireRole("admin", auth.HandleAdminTemplatesNew))
	mux.HandleFunc("POST /org/{orgID}/admin/templates", auth.RequireRole("admin", auth.HandleAdminTemplatesCreate))

	return mux
}
