// dbcheck は PostgreSQL への接続と sqlc 生成クエリの動作を確認するための小さなCLI。
// 主に CI / ローカル検証で「DB接続・マイグレーション適用・クエリ実行」をチェックする。
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

// main はDBに接続し、生成クエリのスモークテストを実行する。
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	dsn := cfg.PostgresURL()
	if dsn == "" {
		logger.Error("missing postgres config")
		os.Exit(1)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("failed to open db", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("db ping failed", "error", err)
		os.Exit(1)
	}

	q := dbgen.New(db)
	one, err := q.Ping(ctx)
	if err != nil {
		logger.Error("sqlc ping failed", "error", err)
		os.Exit(1)
	}
	if one != 1 {
		logger.Error("unexpected ping result", "one", one)
		os.Exit(1)
	}

	row, err := q.UpsertMigrationSmoketest(ctx, 1)
	if err != nil {
		// Likely migrations not applied yet.
		logger.Error("upsert smoketest failed (run migrations first)", "error", err)
		os.Exit(1)
	}

	got, err := q.GetMigrationSmoketest(ctx, row.ID)
	if err != nil {
		logger.Error("get smoketest failed", "error", err)
		os.Exit(1)
	}

	if got.ID != row.ID {
		logger.Error("unexpected row", "row", fmt.Sprintf("%+v", row), "got", fmt.Sprintf("%+v", got))
		os.Exit(1)
	}

	logger.Info("dbcheck ok", "id", got.ID, "created_at", got.CreatedAt)
}
