package tracker

import (
	"encoding/json"
	"log"
	"sync"
	"time"
	"vulnnotes/db"

	"github.com/gorilla/websocket"
)

var (
	clients     = make(map[*websocket.Conn]bool)
	clientsLock sync.RWMutex
	broadcast   = make(chan TaskProgress, 100)
)

// TaskProgress is a single completed CTF task, enriched with session metadata.
type TaskProgress struct {
	SessionID string `json:"sessionId"`
	Nickname  string `json:"nickname"`
	Character string `json:"character"`
	TaskID    int    `json:"taskId"`
	TaskName  string `json:"taskName"`
	Points    int    `json:"points"`
	Timestamp string `json:"timestamp"`
	Details   string `json:"details"`
}

// SessionScore is the leaderboard entry for one player session.
type SessionScore struct {
	SessionID  string         `json:"sessionId"`
	Nickname   string         `json:"nickname"`
	Character  string         `json:"character"`
	SaveCode   string         `json:"saveCode"`
	TotalScore int            `json:"totalScore"`
	Completed  []TaskProgress `json:"completed"`
}

var taskPoints = map[int]int{
	0: 1,
	1: 1, 2: 1, 3: 1,
	4: 2, 5: 2, 6: 2, 7: 2,
	8: 3, 9: 3,
	10: 3,
}

var taskNames = map[int]string{
	0:  "Brute Force",
	1:  "JWT Token Analysis",
	2:  "No Token Expiration",
	3:  "IDOR - Read Others Notes",
	4:  "SQL Injection",
	5:  "Password Disclosure",
	6:  "XSS Attack",
	7:  "IDOR - Delete Others Notes",
	8:  "JWT alg:none Forgery",
	9:  "CSRF Attack",
	10: "🎮 MORTAL KOMBAT Easter Egg",
}

func init() {
	go handleBroadcasts()
}

// CompleteTask records a task completion for a session.
func CompleteTask(sessionID string, taskID int, details string) {
	if db.PG == nil {
		return
	}

	name := taskNames[taskID]
	points := taskPoints[taskID]

	// Check duplicate
	var count int
	db.PG.QueryRow(
		`SELECT COUNT(*) FROM task_completions WHERE session_id = $1 AND task_id = $2`,
		sessionID, taskID,
	).Scan(&count)
	if count > 0 {
		return
	}

	_, err := db.PG.Exec(`
		INSERT INTO task_completions (session_id, task_id, task_name, points, details)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`, sessionID, taskID, name, points, details)
	if err != nil {
		log.Printf("⚠️ failed to save task %d for session %s: %v", taskID, sessionID, err)
		return
	}

	sess, _ := db.GetSessionByID(sessionID)
	nick, char := "", ""
	if sess != nil {
		nick = sess.Nickname
		char = sess.Character
	}

	task := TaskProgress{
		SessionID: sessionID,
		Nickname:  nick,
		Character: char,
		TaskID:    taskID,
		TaskName:  name,
		Points:    points,
		Timestamp: time.Now().Format("15:04:05"),
		Details:   details,
	}

	log.Printf("✅ %s (%s) completed task %d: %s (+%d pts)", nick, char, taskID, name, points)
	broadcast <- task
}

// GetSessionScore returns the score for one session.
func GetSessionScore(sessionID string) SessionScore {
	score := SessionScore{
		SessionID: sessionID,
		Completed: []TaskProgress{},
	}

	if db.PG == nil {
		return score
	}

	sess, _ := db.GetSessionByID(sessionID)
	if sess != nil {
		score.Nickname = sess.Nickname
		score.Character = sess.Character
		score.SaveCode = sess.SaveCode
	}

	rows, err := db.PG.Query(`
		SELECT task_id, task_name, points, details, completed_at
		FROM task_completions WHERE session_id = $1
		ORDER BY completed_at
	`, sessionID)
	if err != nil {
		return score
	}
	defer rows.Close()

	for rows.Next() {
		var t TaskProgress
		var ts time.Time
		rows.Scan(&t.TaskID, &t.TaskName, &t.Points, &t.Details, &ts)
		t.Timestamp = ts.Format("15:04:05")
		t.SessionID = sessionID
		t.Nickname = score.Nickname
		t.Character = score.Character
		score.TotalScore += t.Points
		score.Completed = append(score.Completed, t)
	}
	return score
}

// GetAllScores returns the leaderboard (all sessions, ordered by score).
func GetAllScores() []SessionScore {
	if db.PG == nil {
		return []SessionScore{}
	}

	rows, err := db.PG.Query(`
		SELECT s.id, s.save_code, s.nickname, s.character,
		       COALESCE(SUM(tc.points), 0) AS total_score
		FROM sessions s
		LEFT JOIN task_completions tc ON tc.session_id = s.id
		GROUP BY s.id, s.save_code, s.nickname, s.character
		ORDER BY total_score DESC, s.created_at ASC
	`)
	if err != nil {
		return []SessionScore{}
	}
	defer rows.Close()

	scores := []SessionScore{}
	for rows.Next() {
		var s SessionScore
		rows.Scan(&s.SessionID, &s.SaveCode, &s.Nickname, &s.Character, &s.TotalScore)
		s.Completed = []TaskProgress{}
		scores = append(scores, s)
	}
	return scores
}

func RegisterClient(conn *websocket.Conn) {
	clientsLock.Lock()
	clients[conn] = true
	clientsLock.Unlock()

	msg := map[string]any{
		"type":   "init",
		"scores": GetAllScores(),
	}
	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

func UnregisterClient(conn *websocket.Conn) {
	clientsLock.Lock()
	delete(clients, conn)
	clientsLock.Unlock()
	conn.Close()
}

func handleBroadcasts() {
	for task := range broadcast {
		msg := map[string]any{
			"type": "task_completed",
			"task": task,
		}
		data, _ := json.Marshal(msg)

		var failed []*websocket.Conn
		clientsLock.RLock()
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
				client.Close()
				failed = append(failed, client)
			}
		}
		clientsLock.RUnlock()

		if len(failed) > 0 {
			clientsLock.Lock()
			for _, c := range failed {
				delete(clients, c)
			}
			clientsLock.Unlock()
		}
	}
}
