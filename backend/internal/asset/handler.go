package asset

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var repo *Repository

func RegisterRoutes(mux *http.ServeMux, r *Repository) {
	repo = r
	mux.HandleFunc("GET /api/v1/assets", listAssets)
	mux.HandleFunc("POST /api/v1/assets", createAsset)
	mux.HandleFunc("GET /api/v1/assets/{id}", getAsset)
	mux.HandleFunc("PUT /api/v1/assets/{id}", updateAsset)
	mux.HandleFunc("DELETE /api/v1/assets/{id}", deleteAsset)
	mux.HandleFunc("GET /api/v1/asset-types", listTypes)
}

func listAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := repo.ListAssets(r.Context())
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if assets == nil {
		assets = []Asset{}
	}
	respondOK(w, assets)
}

func createAsset(w http.ResponseWriter, r *http.Request) {
	var in CreateAssetInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if in.Name == "" {
		respondErr(w, http.StatusBadRequest, "name is required")
		return
	}
	if in.Status == "" {
		in.Status = "online"
	}
	id, err := repo.CreateAsset(r.Context(), in)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	created, _ := repo.GetAsset(r.Context(), id)
	if created == nil {
		respondErr(w, http.StatusInternalServerError, "failed to read created asset")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "ok", "data": created})
}

func getAsset(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := repo.GetAsset(r.Context(), id)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if a == nil {
		respondErr(w, http.StatusNotFound, "not found")
		return
	}
	respondOK(w, a)
}

func updateAsset(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in UpdateAssetInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := repo.UpdateAsset(r.Context(), id, in); err != nil {
		respondErr(w, http.StatusNotFound, "not found")
		return
	}
	a, _ := repo.GetAsset(r.Context(), id)
	respondOK(w, a)
}

func deleteAsset(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := repo.DeleteAsset(r.Context(), id); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w, map[string]string{"status": "deleted"})
}

func listTypes(w http.ResponseWriter, r *http.Request) {
	types, err := repo.ListTypes(r.Context())
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w, types)
}

func parseID(r *http.Request) (uint64, error) {
	return strconv.ParseUint(r.PathValue("id"), 10, 64)
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
