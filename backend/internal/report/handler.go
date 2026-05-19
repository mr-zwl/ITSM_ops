package report

import "net/http"

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/reports/sla", SLAReport)
	mux.HandleFunc("GET /api/v1/reports/capacity", CapacityReport)
	mux.HandleFunc("GET /api/v1/reports/health", HealthReport)
}

func SLAReport(w http.ResponseWriter, r *http.Request)      { stub(w) }
func CapacityReport(w http.ResponseWriter, r *http.Request) { stub(w) }
func HealthReport(w http.ResponseWriter, r *http.Request)   { stub(w) }

func stub(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"code":501,"message":"not implemented"}`))
}
