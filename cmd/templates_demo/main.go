// templates_demo は、WorkflowTemplate の Read Model（workflow_templates）を一気通しで確認する小さなCLI。
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	seed := flag.String("seed", "demo", "issuer/sub seed (deterministic user)")
	flag.Parse()

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

	ctx, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	// Setup org/user/membership inside a short tx
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		logger.Error("begin tx failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = tx.Rollback() }()

	q := dbgen.New(tx)

	org, err := q.CreateOrganization(ctx, "Templates Demo Org")
	if err != nil {
		logger.Error("create org failed", "error", err)
		os.Exit(1)
	}
	orgID := org.ID

	user, err := q.UpsertUserFromOIDC(ctx, dbgen.UpsertUserFromOIDCParams{
		OidcIssuer: "templates-demo",
		OidcSub:    *seed,
		Email:      "templates-demo@example.com",
		Name:       "Templates Demo",
	})
	if err != nil {
		logger.Error("upsert user failed", "error", err)
		os.Exit(1)
	}
	actorUserID := user.ID

	if _, err := q.UpsertMembership(ctx, dbgen.UpsertMembershipParams{OrgID: orgID, UserID: actorUserID, Role: "admin"}); err != nil {
		logger.Error("upsert membership failed", "error", err)
		os.Exit(1)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("tx commit failed", "error", err)
		os.Exit(1)
	}

	q2 := dbgen.New(db)

	definition, _ := json.Marshal(map[string]any{
		"version": 1,
		"steps": []map[string]any{
			{"label": "Approve", "assignee": "admin"},
		},
	})

	row, err := q2.CreateWorkflowTemplate(ctx, dbgen.CreateWorkflowTemplateParams{
		OrgID:           orgID,
		Name:            "Demo Template",
		Description:     "demo",
		Definition:      definition,
		CreatedByUserID: uuid.NullUUID{UUID: actorUserID, Valid: true},
	})
	if err != nil {
		logger.Error("CreateWorkflowTemplate failed (did you run migrations 0008?)", "error", err)
		os.Exit(1)
	}

	list, err := q2.ListWorkflowTemplatesByOrg(ctx, dbgen.ListWorkflowTemplatesByOrgParams{OrgID: orgID, Limit: 50, Offset: 0})
	if err != nil {
		logger.Error("ListWorkflowTemplatesByOrg failed", "error", err)
		os.Exit(1)
	}

	logger.Info("templates", "org_id", orgID.String(), "created_id", row.ID.String(), "count", len(list))

	fmt.Println("OK")
	fmt.Println("org_id:", orgID)
	fmt.Println("user_id:", actorUserID)
	fmt.Println("template_id:", row.ID)
}
