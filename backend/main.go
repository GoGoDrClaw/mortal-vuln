package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"vulnnotes/db"
	"vulnnotes/handlers"
	"vulnnotes/middleware"
	"vulnnotes/tracker"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("📡 WebSocket connection attempt from %s (Origin: %s)", r.RemoteAddr, r.Header.Get("Origin"))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	log.Printf("✅ WebSocket connected: %s", r.RemoteAddr)
	tracker.RegisterClient(conn)
	defer tracker.UnregisterClient(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("🔌 WebSocket disconnected: %v", err)
			break
		}
	}
}

func route(mux *http.ServeMux) {
	// Public — session management
	mux.HandleFunc("/api/session/new",     middleware.CORS(handlers.NewSession))
	mux.HandleFunc("/api/session/restore", middleware.CORS(handlers.RestoreSession))
	mux.HandleFunc("/api/session/check",    middleware.CORS(handlers.CheckSession))
	mux.HandleFunc("/api/session/progress", middleware.CORS(handlers.SessionProgress))
	mux.HandleFunc("/api/characters",       middleware.CORS(handlers.GetCharacters))

	// Public — scores & health
	mux.HandleFunc("/api/health",  middleware.CORS(handlers.Health))
	mux.HandleFunc("/api/scores",  middleware.CORS(handlers.GetScores))

	// Public — flag submission (requires session cookie)
	mux.HandleFunc("/api/flag", middleware.CORS(middleware.RequireSession(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.SubmitFlag(w, r)
		}
	})))

	// Public — reset (requires session cookie, no JWT needed)
	mux.HandleFunc("/api/reset", middleware.CORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.Reset(w, r)
		}
	}))

	// Session-required
	mux.HandleFunc("/api/login", middleware.CORS(middleware.RequireSession(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.Login(w, r)
		}
	})))

	// Auth-required (session + JWT)
	mux.HandleFunc("/api/notes", middleware.CORS(middleware.RequireAuth(handlers.TrackAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetNotes(w, r)
		case http.MethodPost:
			handlers.CreateNote(w, r)
		}
	}))))

	mux.HandleFunc("/api/notes/delete/", middleware.CORS(middleware.RequireAuth(handlers.TrackAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.DeleteNote(w, r)
		}
	}))))

	mux.HandleFunc("/api/notes/", middleware.CORS(middleware.RequireAuth(handlers.TrackAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetNote(w, r)
		}
	}))))

	mux.HandleFunc("/api/admin/users", middleware.CORS(middleware.RequireAuth(handlers.TrackAuth(handlers.AdminUsers))))

	mux.HandleFunc("/api/change-password", middleware.CORS(middleware.RequireAuth(handlers.TrackAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.ChangePassword(w, r)
		}
	}))))

	mux.HandleFunc("/api/toasty", middleware.CORS(middleware.RequireSession(handlers.Toasty)))

	// WebSocket
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		handleWebSocket(w, r)
	})
}

func main() {
	// PostgreSQL — sessions + progress
	if err := db.InitPostgres(); err != nil {
		log.Fatalf("❌ PostgreSQL: %v", err)
	}

	// Ensure /data/sessions/ dir exists
	handlers.InitSessions()

	// Rate limiter GC
	middleware.StartLimiterGC()

	mux := http.NewServeMux()
	route(mux)

	handler := middleware.RateLimitHandler(mux)

	fmt.Printf("🔓 VulnNotes Backend started at %s\n", time.Now().Format("15:04:05"))
	fmt.Printf("   Port: 3000\n")
	fmt.Printf("   WebSocket: ws://localhost:3000/ws\n")

	log.Fatal(http.ListenAndServe(":3000", handler))
}
