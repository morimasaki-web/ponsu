// projector は、プロジェクター（Read Model 投影）の catch-up を手動実行するための小さなCLI。
package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/projector"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	projectorName := flag.String("projector", "demo_aggregate_versions", "projector name")
	orgIDStr := flag.String("org", "", "org id (uuid)")
	batchSize := flag.Int("batch", 100, "batch size")
	flag.Parse()

	if *orgIDStr == "" {
		logger.Error("missing -org")
		os.Exit(2)
	}
	orgID, err := uuid.Parse(*orgIDStr)
	if err != nil {
		logger.Error("invalid -org", "error", err)
		os.Exit(2)
	}

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
	if cfg.PostgresURL() == "" {
		logger.Error("postgres is not configured")
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.PostgresURL())
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

	r := projector.NewDemoAggregateVersionRunner(*projectorName)
	r.BatchSize = int32(*batchSize)
	r.Logger = logger

	ctx, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	total, err := r.CatchUp(ctx, db, orgID)
	if err != nil {
		logger.Error("projector failed", "error", err)
		os.Exit(1)
	}

	logger.Info("projector done", "projector", r.ProjectorName, "org_id", orgID.String(), "total", total)
}
