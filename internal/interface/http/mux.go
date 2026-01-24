// Package http は PonSu のHTTPルーティングと最小のハンドラ群を提供する。
// MVPでは healthz と OIDC 認証関連のエンドポイントを定義する。
package http

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/notify"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/storage"
	gqlapi "github.com/morimasaki-web/ponsu/internal/interface/graphql"
	"github.com/morimasaki-web/ponsu/internal/interface/graphqlctx"
	attachmentsuc "github.com/morimasaki-web/ponsu/internal/usecase/attachments"
)

// NewMux はアプリケーションのHTTPルートを登録した ServeMux を返す。
func NewMux(cfg config.Config, logger *slog.Logger, db *sql.DB) *http.ServeMux {
	if logger == nil {
		logger = slog.Default()
	}

	auth := NewOIDCAuth(cfg, logger, db)
	requestsNotifier := notify.NewSlackNotifier(cfg.SlackWebhookURL)
	publicBaseURL := cfg.PublicBaseURLForLinks()
	gqlServer := gqlapi.NewServer(db, requestsNotifier, publicBaseURL)
	playgroundHandler := gqlapi.PlaygroundHandler("/graphql")

	requestsHandler := NewRequestsHandler(cfg, db, requestsNotifier)
	adminTemplatesHandler := NewAdminTemplatesHandler(db)

	// MVP-070/MVP-071: Attachments (storage backend selectable)
	var attachmentsStorage attachmentsuc.Storage
	attachmentsStorage = storage.NewLocalStorage(cfg.AttachmentsLocalDir)
	if strings.EqualFold(cfg.AttachmentsStorage, "minio") || strings.EqualFold(cfg.AttachmentsStorage, "s3") {
		s3st, err := storage.NewS3Storage(
			context.Background(),
			storage.S3StorageConfig{
				Endpoint:       cfg.AttachmentsS3Endpoint,
				Region:         cfg.AttachmentsS3Region,
				Bucket:         cfg.AttachmentsS3Bucket,
				AccessKey:      cfg.AttachmentsS3AccessKey,
				SecretKey:      cfg.AttachmentsS3SecretKey,
				ForcePathStyle: cfg.AttachmentsS3ForcePathStyle,
			},
		)
		if err != nil {
			logger.Error("failed to init attachments s3 storage; falling back to local", "error", err)
		} else {
			attachmentsStorage = s3st
		}
	}
	attachmentsService := attachmentsuc.Service{DB: db, Storage: attachmentsStorage}

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

	// MVP-050: GraphQL (guarded by login)
	mux.HandleFunc("GET /playground", auth.RequireLogin(func(w http.ResponseWriter, r *http.Request, _ sessionData) {
		playgroundHandler.ServeHTTP(w, r)
	}))

	graphqlHandler := func(w http.ResponseWriter, r *http.Request, sess sessionData) {
		v, err := viewerFromSession(sess)
		if err != nil {
			writeUnauthorized(w, r)
			return
		}
		r = r.WithContext(graphqlctx.WithViewer(r.Context(), v))
		gqlServer.ServeHTTP(w, r)
	}
	mux.HandleFunc("GET /graphql", auth.RequireLogin(graphqlHandler))
	mux.HandleFunc("POST /graphql", auth.RequireLogin(graphqlHandler))

	// MVP-041: Request create (SSR minimal)
	mux.HandleFunc("GET /requests", auth.RequireLogin(requestsHandler.HandleRequestsIndexShortcut))
	mux.HandleFunc("GET /requests/new", auth.RequireLogin(requestsHandler.HandleRequestsNewShortcut))
	mux.HandleFunc("GET /org/{orgID}/requests", auth.RequireLogin(requestsHandler.HandleRequestsIndex))
	mux.HandleFunc("GET /org/{orgID}/requests/new", auth.RequireLogin(requestsHandler.HandleRequestsNew))
	mux.HandleFunc("POST /org/{orgID}/requests", auth.RequireLogin(requestsHandler.HandleRequestsCreate))
	mux.HandleFunc("GET /org/{orgID}/requests/{requestID}", auth.RequireLogin(requestsHandler.HandleRequestsShow))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/submit", auth.RequireLogin(requestsHandler.HandleRequestsSubmit))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/approve", auth.RequireLogin(requestsHandler.HandleRequestsApprove))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/reject", auth.RequireLogin(requestsHandler.HandleRequestsReject))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/return", auth.RequireLogin(requestsHandler.HandleRequestsReturn))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/resubmit", auth.RequireLogin(requestsHandler.HandleRequestsResubmit))

	// MVP-011: RBAC
	mux.HandleFunc("GET /admin/templates/new", auth.RequireRole("admin", adminTemplatesHandler.HandleAdminTemplatesNewShortcut))
	mux.HandleFunc("GET /org/{orgID}/admin/templates", auth.RequireRole("admin", adminTemplatesHandler.HandleAdminTemplatesIndex))
	mux.HandleFunc("GET /org/{orgID}/admin/templates/new", auth.RequireRole("admin", adminTemplatesHandler.HandleAdminTemplatesNew))
	mux.HandleFunc("POST /org/{orgID}/admin/templates", auth.RequireRole("admin", adminTemplatesHandler.HandleAdminTemplatesCreate))

	// MVP-070: Attachments (API)
	mux.HandleFunc("GET /org/{orgID}/requests/{requestID}/attachments", auth.RequireLogin(handleAttachmentsList(attachmentsService)))
	mux.HandleFunc("POST /org/{orgID}/requests/{requestID}/attachments", auth.RequireLogin(handleAttachmentsUpload(attachmentsService, cfg.AttachmentsMaxBytes)))
	mux.HandleFunc("GET /org/{orgID}/requests/{requestID}/attachments/{attachmentID}", auth.RequireLogin(handleAttachmentsDownload(attachmentsService)))

	return mux
}
