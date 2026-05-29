package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	pkgauth "github.com/your-org/itsm_ops/backend/pkg/auth"
)

type User struct {
	ID          uint64 `db:"id"           json:"id"`
	Username    string `db:"username"     json:"username"`
	Password    string `db:"password"     json:"-"`
	DisplayName string `db:"display_name" json:"display_name"`
	Email       string `db:"email"        json:"email"`
	Phone       string `db:"phone"        json:"phone"`
	Status      int    `db:"status"       json:"status"`
	CreatedAt   string `db:"created_at"   json:"created_at"`
	UpdatedAt   string `db:"updated_at"   json:"updated_at"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ctxKey string

const claimsKey ctxKey = "claims"

var (
	database  *sqlx.DB
	jwtSecret string
	jwtExpire time.Duration
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB, secret string, expire time.Duration) {
	database = db
	jwtSecret = secret
	jwtExpire = expire

	mux.HandleFunc("POST /api/v1/auth/login", login)
	mux.HandleFunc("GET /api/v1/auth/me", requireAuth(me))
}

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/healthz" || path == "/readyz" ||
				path == "/api/v1/auth/login" ||
			path == "/api/v1/collect" ||
			path == "/api/v1/ssh" ||
			strings.HasSuffix(path, "/rdp") ||
				!strings.HasPrefix(path, "/api/v1/") {
				next.ServeHTTP(w, r)
				return
			}
			token := extractToken(r)
			if token == "" {
				respondErr(w, http.StatusUnauthorized, "missing token")
				return
			}
			claims, err := pkgauth.ParseToken(secret, token)
			if err != nil {
				respondErr(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Username == "" || req.Password == "" {
		respondErr(w, http.StatusBadRequest, "username and password required")
		return
	}
	var user User
	err := database.GetContext(r.Context(), &user, "SELECT * FROM users WHERE username = ?", req.Username)
	if err != nil {
		respondErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !pkgauth.CheckPassword(user.Password, req.Password) {
		respondErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := pkgauth.GenerateToken(jwtSecret, user.ID, user.Username, jwtExpire)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "token generation failed")
		return
	}
	respondOK(w, map[string]any{"token": token, "user": user})
}

func me(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(claimsKey).(*pkgauth.Claims)
	var user User
	err := database.GetContext(r.Context(), &user, "SELECT * FROM users WHERE id = ?", claims.UserID)
	if err != nil {
		respondErr(w, http.StatusNotFound, "user not found")
		return
	}
	respondOK(w, user)
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			respondErr(w, http.StatusUnauthorized, "missing token")
			return
		}
		claims, err := pkgauth.ParseToken(jwtSecret, token)
		if err != nil {
			respondErr(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

func extractToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return h[7:]
	}
	return ""
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
