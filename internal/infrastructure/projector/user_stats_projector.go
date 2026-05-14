package projector

import (
	"context"

	"github.com/morimasaki-web/ponsu/internal/domain/request"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

const UserStatsProjectorName = "user_stats"

// UserStatsProjector は request.* イベントを user_stats/read model に同期投影する。
func UserStatsProjector() Runner {
	return Runner{
		ProjectorName: UserStatsProjectorName,
		BatchSize:     200,
		Apply: func(ctx context.Context, q *dbgen.Queries, e dbgen.ListEventsForProjectorRow) error {
			actorUserID, err := metadataActorUserID(e.Metadata)
			if err != nil {
				return err
			}
			// actor_user_id が無効な場合は無視（システムイベント等）
			if !actorUserID.Valid {
				return nil
			}

			switch e.EventType {
			case request.EventTypeCreated:
				// user_statsを更新
				return q.UpsertUserStatsOnRequestCreated(ctx, dbgen.UpsertUserStatsOnRequestCreatedParams{
					UserID:        actorUserID.UUID,
					RequestCount:  1,
					LastRequestAt: e.OccurredAt,
				})
			default:
				// ignore unrelated events
				return nil
			}
		},
	}
}
