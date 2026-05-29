package alert

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type AlertRule struct {
	ID          uint64  `db:"id"           json:"id"`
	Name        string  `db:"name"         json:"name"`
	MetricCode  string  `db:"metric_code"  json:"metric_code"`
	ConditionOp string  `db:"condition_op" json:"condition_op"`
	Threshold   float64 `db:"threshold"    json:"threshold"`
	Severity    string  `db:"severity"     json:"severity"`
	DurationSec int     `db:"duration" json:"duration_sec"`
	Extra       *string `db:"extra"        json:"extra"`
	Enabled     int     `db:"enabled"      json:"enabled"`
	CreatedAt   string  `db:"created_at"   json:"created_at"`
	UpdatedAt   string  `db:"updated_at"   json:"updated_at"`
}

type AlertEvent struct {
	ID         uint64  `db:"id"          json:"id"`
	RuleID     uint64  `db:"rule_id"     json:"rule_id"`
	AssetID    *uint64 `db:"asset_id"    json:"asset_id"`
	Severity   string  `db:"severity"    json:"severity"`
	Message    string  `db:"message"     json:"message"`
	CurrentVal float64 `db:"current_val" json:"current_val"`
	Status     string  `db:"status"      json:"status"`
	FiredAt    string  `db:"fired_at"    json:"fired_at"`
	ResolvedAt *string `db:"resolved_at" json:"resolved_at"`
	AckedBy    *uint64 `db:"acked_by"    json:"acked_by"`
	AckedAt    *string `db:"acked_at"    json:"acked_at"`
}

func RegisterRoutes(mux *http.ServeMux, database *sqlx.DB) {
	db = database
	mux.HandleFunc("GET /api/v1/alert-rules", listRules)
	mux.HandleFunc("GET /api/v1/alert-events", listEvents)
	mux.HandleFunc("POST /api/v1/alert-events/{id}/ack", ackEvent)
}

func listRules(w http.ResponseWriter, r *http.Request) {
	var rules []AlertRule
	if err := db.SelectContext(r.Context(), &rules, "SELECT * FROM alert_rules ORDER BY id"); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w, rules)
}

func listEvents(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	query := "SELECT * FROM alert_events"
	args := []any{}
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY fired_at DESC LIMIT 100"
	var events []AlertEvent
	if err := db.SelectContext(r.Context(), &events, query, args...); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if events == nil {
		events = []AlertEvent{}
	}
	respondOK(w, events)
}

func ackEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := db.ExecContext(r.Context(),
		"UPDATE alert_events SET status='acked', acked_at=NOW() WHERE id=? AND status='firing'", id)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w, map[string]string{"status": "acked"})
}

func Evaluate(database *sqlx.DB) {
	var rules []AlertRule
	if err := database.Select(&rules, "SELECT * FROM alert_rules WHERE enabled=1"); err != nil {
		slog.Error("alert eval: load rules", "error", err)
		return
	}

	for _, rule := range rules {
		var rows []struct {
			AssetID uint64  `db:"asset_id"`
			Value   float64 `db:"value"`
		}
		query := `SELECT asset_id, value FROM metric_data
			WHERE metric_code=? AND collected_at > DATE_SUB(NOW(), INTERVAL 2 MINUTE)
			ORDER BY collected_at DESC`
		if err := database.Select(&rows, query, rule.MetricCode); err != nil {
			slog.Error("alert eval: query metric_data", "rule", rule.Name, "error", err)
			continue
		}
		slog.Info("alert eval: checked rule", "rule", rule.Name, "metric", rule.MetricCode, "rows", len(rows), "op", rule.ConditionOp, "threshold", rule.Threshold)

		seen := map[uint64]bool{}
		for _, row := range rows {
			if seen[row.AssetID] {
				continue
			}
			seen[row.AssetID] = true

			slog.Info("alert eval: comparing", "asset_id", row.AssetID, "value", row.Value, "op", rule.ConditionOp, "threshold", rule.Threshold)
			triggered := false
			switch rule.ConditionOp {
			case ">":
				triggered = row.Value > rule.Threshold
			case ">=":
				triggered = row.Value >= rule.Threshold
			case "<":
				triggered = row.Value < rule.Threshold
			case "<=":
				triggered = row.Value <= rule.Threshold
			}

			if !triggered {
				continue
			}
			slog.Warn("alert eval: threshold triggered", "rule", rule.Name, "asset_id", row.AssetID, "value", row.Value, "threshold", rule.Threshold)

			var existing int
			database.Get(&existing,
				"SELECT COUNT(*) FROM alert_events WHERE rule_id=? AND asset_id=? AND status='firing'",
				rule.ID, row.AssetID)
			if existing > 0 {
				slog.Info("alert eval: already firing, skip", "rule_id", rule.ID, "asset_id", row.AssetID)
				continue
			}

			msg := fmt.Sprintf("%s: %.2f %s %.2f", rule.Name, row.Value, rule.ConditionOp, rule.Threshold)
			assetID := row.AssetID
			database.Exec(
				`INSERT INTO alert_events (rule_id, asset_id, severity, message, current_val, status)
				 VALUES (?, ?, ?, ?, ?, 'firing')`,
				rule.ID, assetID, rule.Severity, msg, row.Value)

			slog.Warn("alert fired", "rule", rule.Name, "asset_id", assetID, "value", row.Value, "threshold", rule.Threshold)
		}
	}
}

func respondOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "ok", "data": data})
}

func respondErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": status, "message": msg})
}
