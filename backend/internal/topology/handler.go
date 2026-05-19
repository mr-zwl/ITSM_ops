package topology

import "net/http"

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/topology", GetTopology)
	mux.HandleFunc("GET /api/v1/topology/links", ListLinks)
}

func GetTopology(w http.ResponseWriter, r *http.Request) { stub(w) }
func ListLinks(w http.ResponseWriter, r *http.Request)   { stub(w) }

func stub(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"code":501,"message":"not implemented"}`))
}
