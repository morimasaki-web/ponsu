package projector

import (
	"context"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

// NewDemoAggregateVersionRunner は、全イベントを対象に "aggregate ごとの最新 version" を記録する
// デモ用プロジェクターを返す。
//
// 実運用では event_type ごとに Read Model を更新する projector を追加していく。
func NewDemoAggregateVersionRunner(projectorName string) Runner {
	return Runner{
		ProjectorName: projectorName,
		BatchSize:     100,
		Apply: func(ctx context.Context, q *dbgen.Queries, e dbgen.ListEventsForProjectorRow) error {
			_, err := q.UpsertDemoAggregateVersion(ctx, dbgen.UpsertDemoAggregateVersionParams{
				OrgID:         e.OrgID,
				AggregateType: e.AggregateType,
				AggregateID:   e.AggregateID,
				LastVersion:   e.Version,
			})
			return err
		},
	}
}
