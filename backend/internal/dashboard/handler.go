package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type DashboardStats struct {
	TotalAssets    int `json:"total_assets"`
	OnlineAssets   int `json:"online_assets"`
	FiredAlerts    int `json:"fired_alerts"`
	CriticalAlerts int `json:"critical_alerts"`
	WarningAlerts  int `json:"warning_alerts"`
}

type LatestMetric struct {
	AssetID   uint64   `json:"asset_id"`
	AssetName string   `json:"asset_name"`
	AssetIP   string   `json:"asset_ip"`
	CPUUsage  *float64 `json:"cpu_usage"`
	MemUsage  *float64 `json:"mem_usage"`
	DiskUsage *float64 `json:"disk_usage"`
}

type RecentAlert struct {
	ID         uint64   `json:"id"`
	AssetID    *uint64  `json:"asset_id"`
	Severity   string   `json:"severity"`
	Message    string   `json:"message"`
	CurrentVal float64  `json:"current_val"`
	Status     string   `json:"status"`
	FiredAt    string   `json:"fired_at"`
}

type DashboardResponse struct {
	Stats   DashboardStats `json:"stats"`
	Metrics []LatestMetric `json:"metrics"`
	Alerts  []RecentAlert  `json:"alerts"`
}

func RegisterRoutes(mux *http.ServeMux, database *sqlx.DB) {
	db = database
	mux.HandleFunc("GET /api/v1/dashboard", getDashboard)
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	resp := DashboardResponse{}

	// Single query for all stats
	_ = db.QueryRowxContext(r.Context(), `
		SELECT
			(SELECT COUNT(*) FROM assets) as total_assets,
			(SELECT COUNT(*) FROM assets WHERE status='online') as online_assets,
			(SELECT COUNT(*) FROM alert_events WHERE status='fired') as fired_alerts,
			(SELECT COUNT(*) FROM alert_events WHERE status='fired' AND severity='critical') as critical_alerts,
			(SELECT COUNT(*) FROM alert_events WHERE status='fired' AND severity='warning') as warning_alerts
	`).Scan(&resp.Stats.TotalAssets, &resp.Stats.OnlineAssets, &resp.Stats.FiredAlerts, &resp.Stats.CriticalAlerts, &resp.Stats.WarningAlerts)

	// Latest metrics per asset using simple join
	type metricRow struct {
		AssetID   uint64   `db:"asset_id"`
		AssetName string   `db:"asset_name"`
		AssetIP   string   `db:"asset_ip"`
		CPUUsage  *float64 `db:"cpu_usage"`
		MemUsage  *float64 `db:"mem_usage"`
		DiskUsage *float64 `db:"disk_usage"`
	}

	var mRows []metricRow
	if err := db.SelectContext(r.Context(), &mRows, `
		SELECT a.id as asset_id, a.name as asset_name, a.ip as asset_ip,
			cpu.val as cpu_usage, mem.val as mem_usage, disk.val as disk_usage
		FROM assets a
		LEFT JOIN (SELECT md.asset_id, md.value as val FROM metric_data md INNER JOIN (SELECT asset_id, MAX(id) as max_id FROM metric_data WHERE metric_code='cpu_usage' GROUP BY asset_id) t ON md.id = t.max_id) cpu ON a.id = cpu.asset_id
		LEFT JOIN (SELECT md.asset_id, md.value as val FROM metric_data md INNER JOIN (SELECT asset_id, MAX(id) as max_id FROM metric_data WHERE metric_code='mem_usage' GROUP BY asset_id) t ON md.id = t.max_id) mem ON a.id = mem.asset_id
		LEFT JOIN (SELECT md.asset_id, md.value as val FROM metric_data md INNER JOIN (SELECT asset_id, MAX(id) as max_id FROM metric_data WHERE metric_code='disk_usage' GROUP BY asset_id) t ON md.id = t.max_id) disk ON a.id = disk.asset_id
		ORDER BY a.id
	`); err == nil {
		for _, mr := range mRows {
			resp.Metrics = append(resp.Metrics, LatestMetric{
				AssetID:   mr.AssetID,
				AssetName: mr.AssetName,
				AssetIP:   mr.AssetIP,
				CPUUsage:  mr.CPUUsage,
				MemUsage:  mr.MemUsage,
				DiskUsage: mr.DiskUsage,
			})
		}
	}

	// Recent alerts
	_ = db.SelectContext(r.Context(), &resp.Alerts, "SELECT id, asset_id, severity, message, current_val, status, fired_at FROM alert_events ORDER BY fired_at DESC LIMIT 20")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": resp, "message": "ok"})
}
