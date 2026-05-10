package api

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/YumikoKawaii/angelix/pkg/apitypes"
	"github.com/YumikoKawaii/angelix/server/model"
	"github.com/YumikoKawaii/angelix/server/otlp"
)

func (s *Server) handleOTELTraces(w http.ResponseWriter, r *http.Request) {
	memberID := memberIDFromCtx(r.Context())

	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	spans, err := otlp.ParseTraces(memberID, body)
	if err != nil {
		slog.Warn("otlp parse error", "member_id", memberID, "err", err)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
		return
	}

	if len(spans) > 0 {
		if err := s.metrics.RecordSpans(memberID, spans); err != nil {
			slog.Error("record spans", "member_id", memberID, "err", err)
		} else {
			slog.Debug("recorded spans", "member_id", memberID, "count", len(spans))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

func (s *Server) handleOTELMetrics(w http.ResponseWriter, r *http.Request) {
	s.ackOTEL(w, r, model.EventMetrics)
}

func (s *Server) handleOTELLogs(w http.ResponseWriter, r *http.Request) {
	s.ackOTEL(w, r, model.EventLogs)
}

func (s *Server) ackOTEL(w http.ResponseWriter, r *http.Request, t model.EventType) {
	memberID := memberIDFromCtx(r.Context())
	n, _ := io.Copy(io.Discard, io.LimitReader(r.Body, 10<<20))
	slog.Debug("otel signal received", "type", t, "member_id", memberID, "bytes", n)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	memberID := memberIDFromCtx(r.Context())
	s.writeMetricsSummary(w, memberID)
}

func (s *Server) handleAdminGetMetrics(w http.ResponseWriter, r *http.Request) {
	s.writeMetricsSummary(w, r.PathValue("id"))
}

func (s *Server) writeMetricsSummary(w http.ResponseWriter, memberID string) {
	summary, err := s.metrics.GetSpanSummary(memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load metrics")
		return
	}
	tools := make([]apitypes.ToolStat, 0, len(summary.TopTools))
	for _, t := range summary.TopTools {
		tools = append(tools, apitypes.ToolStat{
			Name: t.Name, Count: t.Count, AvgMs: t.AvgMs, ErrorRate: t.ErrorRate,
		})
	}
	writeJSON(w, http.StatusOK, apitypes.MetricsSummaryResponse{
		MemberID:   summary.MemberID,
		TotalSpans: summary.TotalSpans,
		ErrorSpans: summary.ErrorSpans,
		TopTools:   tools,
	})
}
