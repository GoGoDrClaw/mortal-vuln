package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"vulnnotes/db"
	"vulnnotes/models"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
)

var (
	SessionDBs     = map[string]*sql.DB{}
	SessionDBsLock sync.RWMutex
)

var JWTSecret = []byte(envOr("JWT_SECRET", "secret123"))

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// GetSessionDB returns (or opens) the isolated SQLite database for a session.
// The DB is only opened — seeding is done separately by handlers.
func GetSessionDB(sessionID string) *sql.DB {
	if sessionID == "" {
		return nil
	}

	SessionDBsLock.RLock()
	d, ok := SessionDBs[sessionID]
	SessionDBsLock.RUnlock()
	if ok {
		return d
	}

	dbPath := fmt.Sprintf("/data/sessions/%s.db", sessionID)
	d, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil
	}

	SessionDBsLock.Lock()
	SessionDBs[sessionID] = d
	SessionDBsLock.Unlock()
	return d
}

// RequireSession validates the "session_id" cookie against PostgreSQL.
func RequireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_id")
		if err != nil || c.Value == "" {
			writeJSON(w, 401, map[string]string{"error": "No session — start or restore a game first"})
			return
		}

		sess, err := db.GetSessionByID(c.Value)
		if err != nil || sess == nil {
			writeJSON(w, 401, map[string]string{"error": "Invalid session"})
			return
		}

		r.Header.Set("X-Session-ID", sess.ID)
		r.Header.Set("X-Character", sess.Character)
		r.Header.Set("X-Nickname", sess.Nickname)
		next(w, r)
	}
}

func Authenticate(r *http.Request) (*models.Claims, error) {
	tokenStr := ""
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		tokenStr = strings.TrimPrefix(auth, "Bearer ")
	} else if c, err := r.Cookie("token"); err == nil {
		tokenStr = c.Value
	}
	if tokenStr == "" {
		return nil, fmt.Errorf("no token")
	}

	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		switch t.Method.(type) {
		case *jwt.SigningMethodHMAC:
			return JWTSecret, nil
		default:
			return jwt.UnsafeAllowNoneSignatureType, nil
		}
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}
	return claims, nil
}

// RequireAuth wraps RequireSession + JWT validation.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return RequireSession(func(w http.ResponseWriter, r *http.Request) {
		claims, err := Authenticate(r)
		if err != nil {
			writeJSON(w, 401, map[string]string{"error": "Unauthorized: " + err.Error()})
			return
		}
		r.Header.Set("X-User-ID", strconv.Itoa(claims.ID))
		r.Header.Set("X-Username", claims.Username)
		r.Header.Set("X-Role", claims.Role)
		next(w, r)
	})
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
