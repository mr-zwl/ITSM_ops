package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/your-org/itsm_ops/backend/internal/alert"
	"github.com/your-org/itsm_ops/backend/internal/asset"
	"github.com/your-org/itsm_ops/backend/internal/auth"
	"github.com/your-org/itsm_ops/backend/internal/config"
	"github.com/your-org/itsm_ops/backend/internal/metric"
	"github.com/your-org/itsm_ops/backend/internal/middleware"
	"github.com/your-org/itsm_ops/backend/internal/notify"
	"github.com/your-org/itsm_ops/backend/internal/report"
	"github.com/your-org/itsm_ops/backend/internal/topology"
)

type RouterOptions struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *sqlx.DB
}

func NewRouter(opts RouterOptions) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "itsm-ops-backend"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /api/v1/system/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"name": "ITSM_ops", "env": opts.Config.App.Env})
	})

	auth.RegisterRoutes(mux, opts.DB, opts.Config.JWT.Secret, opts.Config.JWT.Expire)
	asset.RegisterRoutes(mux, asset.NewRepository(opts.DB))
	metric.RegisterRoutes(mux, opts.DB)
	alert.RegisterRoutes(mux, opts.DB)
	notify.RegisterRoutes(mux, opts.DB)
	topology.RegisterRoutes(mux)
	report.RegisterRoutes(mux)

	var handler http.Handler = mux
	handler = auth.AuthMiddleware(opts.Config.JWT.Secret)(handler)
	handler = middleware.NewRateLimiter(100, 60*time.Second).Middleware(handler)
	handler = loggingMiddleware(opts.Logger, handler)

	return handler
}

func loggingMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		log.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(startedAt).String())
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload map[string]string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
