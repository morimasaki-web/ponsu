// PonSu のHTTPサーバを起動するエントリポイント。
// 設定読み込み・ルーティング設定・Graceful Shutdown を担当する。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	web "github.com/morimasaki-web/ponsu/internal/interface/http"
)

// main は HTTP サーバを起動し、シグナル受信で安全に停止する。
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           web.NewMux(cfg, logger),
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
