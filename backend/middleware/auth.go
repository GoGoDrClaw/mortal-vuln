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
	"vulnnotes/models"

	"github.com/golang-jwt/jwt/v5"
)

var (
	Teams   = getTeams()
	DBs     = map[string]*sql.DB{}
	DBsLock sync.RWMutex
)

func getTeams() []string {
	// All 23 UMK3 characters ordered by popularity
	allTeams := []string{
		"scorpion",       // 🦂 most iconic
		"subzero",        // ❄️  most iconic
		"liukang",        // 🔥 main protagonist
		"kitana",         // 👸 fan favourite
		"raiden",         // ⚡ fan favourite
		"jax",            // 💪
		"mileena",        // 🎭
		"kunglao",        // 🎩
		"sonya",          // 🎖️
		"shangtsung",     // 💀
		"kano",           // 🔴
		"nightwolf",      // 🐺
		"cyrax",          // 🤖
		"sektor",         // 🔴
		"kabal",          // ⚔️
		"jade",           // 💚
		"sindel",         // 👑
		"ermac",          // 👻
		"sheeva",         // 👊
		"stryker",        // 🚔
		"smoke",          // 💨
		"noobsaibot",     // 🌑
	}

	count := 6 // default
	if s := os.Getenv("TEAM_COUNT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 2 && n <= len(allTeams) {
			count = n
		}
	}

	return allTeams[:count]
}

var JWTSecret = []byte("secret123")

func ValidTeam(team string) bool {
	for _, t := range Teams {
		if t == team {
			return true
		}
	}
	return false
}

func GetTeamDB(team string) *sql.DB {
	DBsLock.RLock()
	defer DBsLock.RUnlock()
	return DBs[team]
}

func TeamFromRequest(r *http.Request) string {
	if c, err := r.Cookie("team"); err == nil && ValidTeam(c.Value) {
		return c.Value
	}
	return ""
}

func RequireTeam(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		team := TeamFromRequest(r)
		if team == "" {
			writeJSON(w, 400, map[string]string{"error": "No team selected"})
			return
		}
		r.Header.Set("X-Team", team)
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

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return RequireTeam(func(w http.ResponseWriter, r *http.Request) {
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
		// Allow requests from any subdomain of our domain (for local/remote setup)
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
