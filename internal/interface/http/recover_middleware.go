package http

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// RecoverMiddleware は、panic発生時にサーバを落とさず500エラーを返すHTTPミドルウェアです。
//
// ・panic発生時、recoverで捕捉し、スタックトレースをslog.Loggerで出力します。
// ・クライアントにはwriteInternalErrorで500エラーを返し、詳細はレスポンスに含めません。
func RecoverMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if next == nil {
		next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeInternalError(w, r, logger, errors.New("nil next handler"))
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error(
					"panic recovered",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", rec,
					"stack", string(debug.Stack()),
				)
				// レスポンスには詳細を出さない（情報漏えい防止）。ログは上で出している。
				writeInternalError(w, r, nil, nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
