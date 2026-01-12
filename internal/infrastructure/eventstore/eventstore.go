// Package eventstore は event_store テーブルへのアクセス（append/load）を提供する。
// MVP-021 では、楽観ロック（expected version）をDB側で検証して安全にイベント追記できることを重視する。
package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

// ErrVersionConflict は、期待した version とDB上の最新 version が一致せず append が拒否されたことを表す。
var ErrVersionConflict = errors.New("eventstore version conflict")

// Append は aggregate の次の version にイベントを1件追記する。
//
// expectedVersion は「追記前の最新 version」を指定する（新規は 0）。
// dbtx には *sql.DB または *sql.Tx を渡せるため、同一TX内で Read Model 更新も行える。
func Append(
	ctx context.Context,
	dbtx dbgen.DBTX,
	orgID uuid.UUID,
	aggregateType string,
	aggregateID uuid.UUID,
	expectedVersion int32,
	eventType string,
	payload json.RawMessage,
	metadata json.RawMessage,
) (dbgen.AppendEventRow, error) {
	q := dbgen.New(dbtx)
	row, err := q.AppendEvent(ctx, dbgen.AppendEventParams{
		Column1: orgID,
		Column2: aggregateType,
		Column3: aggregateID,
		Column4: expectedVersion + 1,
		Column5: eventType,
		Column6: payload,
		Column7: metadata,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbgen.AppendEventRow{}, ErrVersionConflict
		}
		return dbgen.AppendEventRow{}, err
	}
	return row, nil
}

// GetVersion は aggregate の最新 version（イベント未作成なら 0）を返す。
func GetVersion(
	ctx context.Context,
	dbtx dbgen.DBTX,
	orgID uuid.UUID,
	aggregateType string,
	aggregateID uuid.UUID,
) (int32, error) {
	q := dbgen.New(dbtx)
	return q.GetAggregateVersion(ctx, dbgen.GetAggregateVersionParams{
		OrgID:         orgID,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
	})
}

// ListByAggregate は aggregate のイベントを version の昇順で返す。
func ListByAggregate(
	ctx context.Context,
	dbtx dbgen.DBTX,
	orgID uuid.UUID,
	aggregateType string,
	aggregateID uuid.UUID,
	fromVersion int32,
) ([]dbgen.ListEventsByAggregateRow, error) {
	q := dbgen.New(dbtx)
	return q.ListEventsByAggregate(ctx, dbgen.ListEventsByAggregateParams{
		OrgID:         orgID,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Version:       fromVersion,
	})
}
