package metric

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type MetricPoint struct {
	AssetID     uint64  `json:"asset_id"`
	MetricCode  string  `json:"metric_code"`
	Value       float64 `json:"value"`
	CollectedAt string  `json:"collected_at,omitempty"`
}

type IngestRequest struct {
	Points []MetricPoint `json:"points"`
}

type MetricDef struct {
	ID          uint64 `db:"id"          json:"id"`
	Code        string `db:"code"        json:"code"`
	Name        string `db:"name"        json:"name"`
	Unit        string `db:"unit"        json:"unit"`
	DataType    string `db:"data_type"   json:"data_type"`
	Description string `db:"description" json:"description"`
	CreatedAt   string `db:"created_at"  json:"created_at"`
}

type MetricData struct {
	ID          uint64  `db:"id"           json:"id"`
	AssetID     uint64  `db:"asset_id"     json:"asset_id"`
	MetricCode  string  `db:"metric_code"  json:"metric_code"`
	Value       float64 `db:"value"        json:"value"`
	CollectedAt string  `db:"collected_at" json:"collected_at"`
	CreatedAt   string  `db:"created_at"   json:"created_at"`
}

func RegisterRoutes(mux *http.ServeMux, database *sqlx.DB) {
	db = database
	mux.HandleFunc("GET /api/v1/metrics", listDefinitions)
	mux.HandleFunc("POST /api/v1/collect", ingest)
	mux.HandleFunc("GET /api/v1/metric-data", queryData)
}

func listDefinitions(w http.ResponseWriter, r *http.Request) {
	var defs []MetricDef
	if err := db.SelectContext(r.Context(), &defs, "SELECT * FROM metric_definitions ORDER BY id"); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w, defs)
}

func ingest(w http.ResponseWriter, r *http.Request) {
	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.Points) == 0 {
		respondErr(w, http.StatusBadRequest, "empty points")
		return
	}

	tx, err := db.BeginTxx(r.Context(), nil)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	inserted := 0
	for _, p := range req.Points {
		if p.AssetID == 0 || p.MetricCode == "" {
			continue
		}
		ts := p.CollectedAt
		if ts == "" {
			ts = ""
			_, err = tx.ExecContext(r.Context(),
				"INSERT INTO metric_data (asset_id, metric_code, value, collected_at) VALUES (?, ?, ?, NOW())",
				p.AssetID, p.MetricCode, p.Value)
		} else {
			_, err = tx.ExecContext(r.Context(),
				"INSERT INTO metric_data (asset_id, metric_code, value, collected_at) VALUES (?, ?, ?, ?)",
				p.AssetID, p.MetricCode, p.Value, ts)
		}
		if err != nil {
			_ = tx.Rollback()
			respondErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondOK(w, map[string]int{"inserted": inserted})
}

func queryData(w http.ResponseWriter, r *http.Request) {
	assetID := r.URL.Query().Get("asset_id")
	code := r.URL.Query().Get("metric_code")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "100"
	}

	query := "SELECT * FROM metric_data WHERE 1=1"
	args := []any{}
	if assetID != "" {
		query += " AND asset_id = ?"
		args = append(args, assetID)
	}
	if code != "" {
		query += " AND metric_code = ?"
		args = append(args, code)
	}
	query += " ORDER BY collected_at DESC LIMIT ?"
	args = append(args, limit)

	var data []MetricData
	if err := db.SelectContext(r.Context(), &data, query, args...); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if data == nil {
		data = []MetricData{}
	}
	respondOK(w, data)
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
