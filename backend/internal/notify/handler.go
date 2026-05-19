package notify

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type Channel struct {
	ID        uint64 `db:"id"         json:"id"`
	Name      string `db:"name"       json:"name"`
	Type      string `db:"type"       json:"type"`
	Config    string `db:"config"     json:"config"`
	Enabled   int    `db:"enabled"    json:"enabled"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

type Record struct {
	ID        uint64 `db:"id"         json:"id"`
	ChannelID uint64 `db:"channel_id" json:"channel_id"`
	EventID   uint64 `db:"event_id"   json:"event_id"`
	Status    string `db:"status"     json:"status"`
	Message   string `db:"message"    json:"message"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

func RegisterRoutes(mux *http.ServeMux, database *sqlx.DB) {
	db = database
	mux.HandleFunc("GET /api/v1/notify-channels", listChannels)
	mux.HandleFunc("GET /api/v1/notify-records", listRecords)
}

func listChannels(w http.ResponseWriter, r *http.Request) {
	var channels []Channel
	if err := db.SelectContext(r.Context(), &channels, "SELECT * FROM notify_channels ORDER BY id"); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w, channels)
}

func listRecords(w http.ResponseWriter, r *http.Request) {
	var records []Record
	if err := db.SelectContext(r.Context(), &records, "SELECT * FROM notify_records ORDER BY id DESC LIMIT 100"); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if records == nil {
		records = []Record{}
	}
	respondOK(w, records)
}

func SendNotifications(database *sqlx.DB) {
	var events []struct {
		EventID  uint64 `db:"id"`
		Message  string `db:"message"`
		Severity string `db:"severity"`
	}
	err := database.Select(&events,
		`SELECT id, message, severity FROM alert_events
		 WHERE status='firing' AND id NOT IN (SELECT event_id FROM notify_records)`)
	if err != nil || len(events) == 0 {
		return
	}

	var channels []Channel
	if err := database.Select(&channels, "SELECT * FROM notify_channels WHERE enabled=1"); err != nil {
		return
	}

	for _, evt := range events {
		for _, ch := range channels {
			switch ch.Type {
			case "log":
				slog.Warn("NOTIFY", "channel", ch.Name, "severity", evt.Severity, "message", evt.Message)
			}
			database.Exec(
				"INSERT INTO notify_records (channel_id, event_id, status, message) VALUES (?, ?, 'sent', ?)",
				ch.ID, evt.EventID, evt.Message)
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
