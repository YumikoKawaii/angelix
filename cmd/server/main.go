package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"

	"github.com/YumikoKawaii/angelix/server/api"
	"github.com/YumikoKawaii/angelix/server/metrics"
	"github.com/YumikoKawaii/angelix/server/store"
)

var cfg struct {
	AdminToken     string `env:"ANGELIX_ADMIN_TOKEN"     required:"" help:"Admin authentication token"`
	Port           string `env:"PORT"                    default:"8080"       help:"HTTP listen port"`
	DB             string `env:"ANGELIX_DB"              default:"angelix.db" help:"SQLite path for member store"`
	Debug          bool   `env:"DEBUG"                                        help:"Enable debug logging"`
	MetricsBackend string `env:"ANGELIX_METRICS_BACKEND" default:"clickhouse" enum:"clickhouse,duckdb" help:"Metrics storage backend"`

	ClickHouse metrics.ClickHouseConfig `embed:"" prefix:"clickhouse-"`
	DuckDB     metrics.DuckDBConfig     `embed:"" prefix:"duckdb-"`
}

func main() {
	kong.Parse(&cfg,
		kong.Name("angelix-server"),
		kong.Description("Angelix credential and telemetry server"),
		kong.UsageOnError(),
	)

	initLogger(cfg.Debug)

	memberStore, err := store.NewSQLite(cfg.DB)
	if err != nil {
		slog.Error("failed to open member db", "path", cfg.DB, "err", err)
		os.Exit(1)
	}
	defer memberStore.Close()

	metricsStore, err := openMetricsStore()
	if err != nil {
		slog.Error("failed to open metrics store", "backend", cfg.MetricsBackend, "err", err)
		os.Exit(1)
	}
	defer metricsStore.Close()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      api.NewServer(memberStore, metricsStore, cfg.AdminToken),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server started", "addr", srv.Addr, "metrics_backend", cfg.MetricsBackend)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	slog.Info("stopped")
}

func openMetricsStore() (metrics.Store, error) {
	switch cfg.MetricsBackend {
	case "clickhouse":
		slog.Info("metrics backend: clickhouse", "addr", cfg.ClickHouse.Addr, "db", cfg.ClickHouse.Database)
		return metrics.NewClickHouse(cfg.ClickHouse)
	default: // duckdb
		slog.Info("metrics backend: duckdb", "path", cfg.DuckDB.Path)
		return metrics.NewDuckDB(cfg.DuckDB)
	}
}

func initLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}
