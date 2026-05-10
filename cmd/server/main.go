package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YumikoKawaii/angelix/server/api"
	"github.com/YumikoKawaii/angelix/server/metrics"
	"github.com/YumikoKawaii/angelix/server/store"
)

func main() {
	initLogger()

	adminToken := requireEnv("ANGELIX_ADMIN_TOKEN")

	port := envOr("PORT", "8080")
	dbPath := envOr("ANGELIX_DB", "angelix.db")

	memberStore, err := store.NewSQLite(dbPath)
	if err != nil {
		slog.Error("failed to open member db", "path", dbPath, "err", err)
		os.Exit(1)
	}
	defer memberStore.Close()

	metricsStore, err := openMetricsStore()
	if err != nil {
		slog.Error("failed to open metrics store", "err", err)
		os.Exit(1)
	}
	defer metricsStore.Close()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      api.NewServer(memberStore, metricsStore, adminToken),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server started", "addr", srv.Addr)
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

// openMetricsStore reads ANGELIX_METRICS_BACKEND (default: clickhouse) and
// instantiates the appropriate metrics.Store implementation.
func openMetricsStore() (metrics.Store, error) {
	backend := envOr("ANGELIX_METRICS_BACKEND", "clickhouse")

	switch backend {
	case "clickhouse":
		cfg := metrics.ClickHouseConfig{
			Addr:     envOr("CLICKHOUSE_ADDR", "localhost:9000"),
			Database: envOr("CLICKHOUSE_DB", "angelix"),
			Username: envOr("CLICKHOUSE_USER", "default"),
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		}
		slog.Info("metrics backend: clickhouse", "addr", cfg.Addr, "db", cfg.Database)
		return metrics.NewClickHouse(cfg)

	case "duckdb":
		path := envOr("DUCKDB_PATH", "metrics.duckdb")
		slog.Info("metrics backend: duckdb", "path", path)
		return metrics.NewDuckDB(path)

	default:
		return nil, fmt.Errorf("unknown ANGELIX_METRICS_BACKEND %q — use 'clickhouse' or 'duckdb'", backend)
	}
}

func initLogger() {
	level := slog.LevelInfo
	if os.Getenv("DEBUG") == "1" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "%s is required\n", key)
		os.Exit(1)
	}
	return v
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
