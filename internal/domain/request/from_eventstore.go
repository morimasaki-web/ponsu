package request

import (
	"fmt"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

// FromEventStoreRows は sqlc の event_store 行を Request のドメインイベントに変換する。
// aggregate_type の検証は呼び出し側（Usecase/Repo）に任せる。
func FromEventStoreRows(rows []dbgen.ListEventsByAggregateRow) ([]Event, error) {
	events := make([]Event, 0, len(rows))
	for _, r := range rows {
		events = append(events, Event{
			OrgID:      r.OrgID,
			RequestID:  r.AggregateID,
			Version:    r.Version,
			Type:       r.EventType,
			OccurredAt: r.OccurredAt,
			Payload:    r.Payload,
		})
	}
	// Basic sanity: versions should start at 1
	if len(events) > 0 && events[0].Version != 1 {
		return nil, fmt.Errorf("invalid first version: got %d", events[0].Version)
	}
	return events, nil
}
