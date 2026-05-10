package metrics

import "github.com/YumikoKawaii/angelix/server/model"

// Store is the metrics backend interface. Member management stays in store.Store (SQLite).
// Implementations: ClickHouse (default), DuckDB.
type Store interface {
	RecordSpans(memberID string, spans []*model.Span) error
	GetSpanSummary(memberID string) (*model.SpanSummary, error)
	Close() error
}
