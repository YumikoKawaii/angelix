package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/YumikoKawaii/angelix/server/model"
)

const chSchema = `
CREATE TABLE IF NOT EXISTS spans (
    member_id   LowCardinality(String),
    trace_id    String,
    span_id     String,
    name        LowCardinality(String),
    start_time  DateTime64(9, 'UTC'),
    end_time    DateTime64(9, 'UTC'),
    duration_ms Int64,
    is_error    UInt8,
    recorded_at DateTime64(3, 'UTC')
) ENGINE = MergeTree()
ORDER BY (member_id, name, start_time)
PARTITION BY toYYYYMM(start_time)
`

// ClickHouseConfig holds connection parameters.
// Fields carry kong tags so the struct can be embedded directly in a CLI/server config.
type ClickHouseConfig struct {
	Addr     string `env:"CLICKHOUSE_ADDR"     default:"localhost:9000" help:"ClickHouse native address (host:port)"`
	Database string `env:"CLICKHOUSE_DB"       default:"angelix"        help:"ClickHouse database"`
	Username string `env:"CLICKHOUSE_USER"     default:"default"         help:"ClickHouse user"`
	Password string `env:"CLICKHOUSE_PASSWORD" default:""               help:"ClickHouse password"`
}

type ClickHouse struct {
	conn driver.Conn
}

func NewClickHouse(cfg ClickHouseConfig) (*ClickHouse, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:     5 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	if err := conn.Exec(ctx, chSchema); err != nil {
		return nil, fmt.Errorf("clickhouse schema: %w", err)
	}
	return &ClickHouse{conn: conn}, nil
}

func (c *ClickHouse) RecordSpans(memberID string, spans []*model.Span) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	batch, err := c.conn.PrepareBatch(ctx,
		`INSERT INTO spans (member_id, trace_id, span_id, name, start_time, end_time, duration_ms, is_error, recorded_at)`,
	)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	now := time.Now().UTC()
	for _, sp := range spans {
		isErr := uint8(0)
		if sp.IsError {
			isErr = 1
		}
		if err := batch.Append(
			memberID, sp.TraceID, sp.SpanID, sp.Name,
			sp.StartTime, sp.EndTime, sp.DurationMs, isErr, now,
		); err != nil {
			return fmt.Errorf("batch append: %w", err)
		}
	}
	return batch.Send()
}

func (c *ClickHouse) GetSpanSummary(memberID string) (*model.SpanSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := c.conn.Query(ctx, `
		SELECT
			name,
			count()        AS cnt,
			avg(duration_ms) AS avg_ms,
			sum(is_error)  AS errs
		FROM spans
		WHERE member_id = ?
		GROUP BY name
		ORDER BY cnt DESC
	`, memberID)
	if err != nil {
		return nil, fmt.Errorf("query summary: %w", err)
	}
	defer rows.Close()

	summary := &model.SpanSummary{MemberID: memberID}
	for rows.Next() {
		var (
			name   string
			cnt    uint64
			avgMs  float64
			errors uint64
		)
		if err := rows.Scan(&name, &cnt, &avgMs, &errors); err != nil {
			return nil, err
		}
		summary.TotalSpans += int(cnt)
		summary.ErrorSpans += int(errors)
		summary.TopTools = append(summary.TopTools, model.ToolStat{
			Name:      name,
			Count:     int(cnt),
			AvgMs:     avgMs,
			ErrorRate: float64(errors) / float64(cnt),
		})
	}
	return summary, rows.Err()
}

func (c *ClickHouse) Close() error {
	return c.conn.Close()
}
