package projector

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

// ApplyFunc は、イベント1件を Read Model に適用する関数。
// 呼び出し側で TX を開始し、q はその TX に紐づく Queries を渡す。
//
// NOTE: projector はイベントの順序付けとして event_store.global_position を利用する。
// Read Model 側は冪等（同じイベントを複数回適用しても壊れない）にすること。

type ApplyFunc func(ctx context.Context, q *dbgen.Queries, e dbgen.ListEventsForProjectorRow) error

// Runner は、event_store のイベントを順に読み込み、Read Model に投影してチェックポイントを更新する。
// MVP-022 では「同期投影（catch-up）」の最小基盤を提供する。
//
// 併走制御（排他）はMVPでは行わない。必要なら将来、advisory lock などを追加する。
//
// チェックポイント:
// - projection_checkpoints.last_position は、最後に適用済みの event_store.global_position。
// - 次回は last_position より大きいイベントを対象にする。

type Runner struct {
	ProjectorName string
	BatchSize     int32
	Apply         ApplyFunc
	Logger        *slog.Logger
}

func (r Runner) batchSizeOrDefault() int32 {
	if r.BatchSize <= 0 {
		return 100
	}
	return r.BatchSize
}

func (r Runner) loggerOrDefault() *slog.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	return slog.Default()
}

// RunOnce は、チェックポイント以降のイベントを最大 BatchSize 件投影する。
// 0 件なら processed=0 を返す。
func (r Runner) RunOnce(ctx context.Context, db *sql.DB, orgID uuid.UUID) (processed int, err error) {
	if db == nil {
		return 0, errors.New("db is nil")
	}
	if r.ProjectorName == "" {
		return 0, errors.New("ProjectorName is required")
	}
	if r.Apply == nil {
		return 0, errors.New("Apply is required")
	}

	logger := r.loggerOrDefault()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback() // no-op if committed
	}()

	q := dbgen.New(tx)

	lastPosition := int64(0)
	cp, err := q.GetProjectionCheckpoint(ctx, dbgen.GetProjectionCheckpointParams{
		OrgID:         orgID,
		ProjectorName: r.ProjectorName,
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	} else {
		lastPosition = cp.LastPosition
	}

	events, err := q.ListEventsForProjector(ctx, dbgen.ListEventsForProjectorParams{
		OrgID:          orgID,
		GlobalPosition: lastPosition,
		Limit:          r.batchSizeOrDefault(),
	})
	if err != nil {
		return 0, err
	}

	if len(events) == 0 {
		if err := tx.Commit(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	for _, e := range events {
		if err := r.Apply(ctx, q, e); err != nil {
			logger.Warn("projector apply failed", "projector", r.ProjectorName, "org_id", orgID.String(), "global_position", e.GlobalPosition, "event_type", e.EventType, "error", err)
			return 0, err
		}
		lastPosition = e.GlobalPosition
	}

	_, err = q.UpsertProjectionCheckpoint(ctx, dbgen.UpsertProjectionCheckpointParams{
		OrgID:         orgID,
		ProjectorName: r.ProjectorName,
		LastPosition:  lastPosition,
	})
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	logger.Info("projector batch applied", "projector", r.ProjectorName, "org_id", orgID.String(), "processed", len(events), "last_position", lastPosition)
	return len(events), nil
}

// CatchUp は、RunOnce を繰り返して 0 件になるまで投影する。
func (r Runner) CatchUp(ctx context.Context, db *sql.DB, orgID uuid.UUID) (total int, err error) {
	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		processed, err := r.RunOnce(ctx, db, orgID)
		if err != nil {
			return total, err
		}
		total += processed
		if processed == 0 {
			return total, nil
		}
	}
}
