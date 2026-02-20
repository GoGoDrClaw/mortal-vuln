package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
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
	mux.HandleFunc("/api/health", middleware.CORS(handlers.Health))
	mux.HandleFunc("/api/select-team", middleware.CORS(handlers.SelectTeam))

	mux.HandleFunc("/api/login", middleware.CORS(middleware.RequireTeam(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.Login(w, r)
		}
	})))

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

	mux.HandleFunc("/api/flag", middleware.CORS(middleware.RequireTeam(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.SubmitFlag(w, r)
		}
	})))

	mux.HandleFunc("/api/reset", middleware.CORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.Reset(w, r)
		}
	}))

	mux.HandleFunc("/api/scores", middleware.CORS(handlers.GetScores))
	mux.HandleFunc("/api/teams", middleware.CORS(handlers.GetTeams))
	mux.HandleFunc("/api/toasty", middleware.CORS(middleware.RequireTeam(handlers.Toasty)))

	// WebSocket endpoint with CORS support
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
	handlers.InitAllDBs()
	mux := http.NewServeMux()
	route(mux)

	fmt.Printf("🔓 VulnNotes Backend started at %s\n", time.Now().Format("15:04:05"))
	fmt.Printf("   Port: 3000\n")
	fmt.Printf("   Teams: %s\n", strings.Join(middleware.Teams, ", "))
	fmt.Printf("   WebSocket: ws://localhost:3000/ws\n")

	log.Fatal(http.ListenAndServe(":3000", mux))
}
