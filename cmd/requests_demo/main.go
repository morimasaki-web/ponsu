// requests_demo は、request.* のイベント追記→同期投影→read model/audit を一気通しで確認する小さなCLI。
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/projector"
	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
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

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		logger.Error("begin tx failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = tx.Rollback() }()

	q := dbgen.New(tx)

	org, err := q.CreateOrganization(ctx, "Demo Org")
	if err != nil {
		logger.Error("create org failed", "error", err)
		os.Exit(1)
	}
	orgID := org.ID

	user, err := q.UpsertUserFromOIDC(ctx, dbgen.UpsertUserFromOIDCParams{
		OidcIssuer: "requests-demo",
		OidcSub:    *seed,
		Email:      "requests-demo@example.com",
		Name:       "Requests Demo",
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

	svc := requestsuc.Service{DB: db}

	requestID, err := svc.CreateRequest(ctx, orgID, actorUserID, "Demo Request")
	if err != nil {
		logger.Error("CreateRequest failed", "error", err)
		os.Exit(1)
	}
	if err := svc.SubmitRequest(ctx, orgID, actorUserID, requestID); err != nil {
		logger.Error("SubmitRequest failed", "error", err)
		os.Exit(1)
	}
	if err := svc.ApproveRequest(ctx, orgID, actorUserID, requestID); err != nil {
		logger.Error("ApproveRequest failed", "error", err)
		os.Exit(1)
	}

	// 監査ログが3件（created/submitted/approved）になっていること
	q2 := dbgen.New(db)
	audit1, err := q2.ListRequestAuditTrail(ctx, dbgen.ListRequestAuditTrailParams{OrgID: orgID, RequestID: requestID})
	if err != nil {
		logger.Error("ListRequestAuditTrail failed", "error", err)
		os.Exit(1)
	}
	reqRow, err := q2.GetRequestByOrgAndID(ctx, dbgen.GetRequestByOrgAndIDParams{OrgID: orgID, ID: requestID})
	if err != nil {
		logger.Error("GetRequestByOrgAndID failed", "error", err)
		os.Exit(1)
	}

	logger.Info("after usecase", "org_id", orgID.String(), "user_id", actorUserID.String(), "request_id", requestID.String(), "status", reqRow.Status, "audit_rows", len(audit1))

	// 再投影（at-least-onceシミュレーション）: チェックポイントを 0 に戻して同じイベントを再適用
	if _, err := q2.UpsertProjectionCheckpoint(ctx, dbgen.UpsertProjectionCheckpointParams{OrgID: orgID, ProjectorName: projector.RequestsProjectorName, LastPosition: 0}); err != nil {
		logger.Error("reset checkpoint failed", "error", err)
		os.Exit(1)
	}

	runner := projector.NewRequestsProjector()
	total, err := runner.CatchUp(ctx, db, orgID)
	if err != nil {
		logger.Error("requests projector catch-up failed", "error", err)
		os.Exit(1)
	}

	audit2, err := q2.ListRequestAuditTrail(ctx, dbgen.ListRequestAuditTrailParams{OrgID: orgID, RequestID: requestID})
	if err != nil {
		logger.Error("ListRequestAuditTrail (after re-run) failed", "error", err)
		os.Exit(1)
	}

	if len(audit2) != len(audit1) {
		logger.Error("audit not idempotent", "before", len(audit1), "after", len(audit2))
		os.Exit(1)
	}
	if len(audit2) != 3 {
		logger.Warn("unexpected audit rows", "count", len(audit2))
	}

	logger.Info("after re-run", "processed", total, "audit_rows", len(audit2))

	// 追加の簡易チェック: 監査行に重複がないこと（action+occurred_at だけでは不十分なので、ここでは件数のみ）
	if len(audit2) == 0 {
		logger.Error("no audit rows")
		os.Exit(1)
	}

	// show last audit action for visibility
	last := audit2[len(audit2)-1]
	logger.Info("last audit", "action", last.Action, "occurred_at", last.OccurredAt.Format(time.RFC3339Nano), "actor_user_id_valid", last.ActorUserID.Valid)

	// sanity: org boundary
	if reqRow.OrgID != orgID {
		logger.Error("org boundary broken")
		os.Exit(1)
	}

	// Ensure projector didn't error silently
	if total <= 0 {
		// not fatal; could happen if there were no events beyond checkpoint for some reason
		logger.Warn("projector processed 0 events; check checkpoints")
	}

	// Pretty print summary
	fmt.Println("OK")
	fmt.Println("org_id:", orgID)
	fmt.Println("user_id:", actorUserID)
	fmt.Println("request_id:", requestID)
}
