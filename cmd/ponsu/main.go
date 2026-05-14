// PonSu のHTTPサーバを起動するエントリポイント。
// 設定読み込み・ルーティング設定・Graceful Shutdown を担当する。
package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	web "github.com/morimasaki-web/ponsu/internal/interface/http"
)

// main は HTTP サーバを起動し、シグナル受信で安全に停止する。
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if loaded, err := config.LoadDotenvLocal(); err != nil {
		logger.Warn("failed to load dotenv", "error", err)
	} else if len(loaded) > 0 {
		logger.Info("loaded dotenv", "files", loaded)
	}
	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	dsn := cfg.PostgresURL()
	var db *sql.DB
	if dsn == "" {
		logger.Warn("postgres is not configured; rbac features disabled")
	} else {
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			logger.Error("failed to open db", "error", err)
			os.Exit(1)
		}
		defer func() { _ = db.Close() }()

		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(pingCtx); err != nil {
			logger.Error("db ping failed", "error", err)
			os.Exit(1)
		}
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           web.NewMux(cfg, logger, db),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("ponsu starting", "addr", server.Addr, "env", cfg.Env)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("ponsu stopped")
}
