package tracker

import (
	"encoding/json"
	"log"
	"sync"
	"time"
	"vulnnotes/middleware"
	"vulnnotes/models"

	"github.com/gorilla/websocket"
)

var (
	progress     = make(map[string]map[int]models.TaskProgress)
	progressLock sync.RWMutex
	clients      = make(map[*websocket.Conn]bool)
	clientsLock  sync.RWMutex
	broadcast    = make(chan models.TaskProgress, 100)
)

var taskPoints = map[int]int{
	0: 1,                    // Brute force
	1: 1, 2: 1, 3: 1,        // Easy
	4: 2, 5: 2, 6: 2, 7: 2,  // Medium
	8: 3, 9: 3,              // Hard
	10: 3,                   // Secret easter egg
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
	for _, team := range middleware.Teams {
		progress[team] = make(map[int]models.TaskProgress)
	}

	// Initialize persistent progress database
	if err := InitProgressDB(); err != nil {
		log.Fatalf("❌ Failed to initialize progress database: %v", err)
	}

	go handleBroadcasts()
}

func CompleteTask(team string, taskID int, details string) {
	progressLock.Lock()
	defer progressLock.Unlock()

	if _, exists := progress[team][taskID]; exists {
		return
	}

	task := models.TaskProgress{
		Team:      team,
		TaskID:    taskID,
		TaskName:  taskNames[taskID],
		Points:    taskPoints[taskID],
		Timestamp: time.Now().Format("15:04:05"),
		Details:   details,
	}

	progress[team][taskID] = task

	// Save to persistent database
	if err := saveTaskToDB(task); err != nil {
		log.Printf("⚠️  Failed to save task to DB: %v", err)
	}

	log.Printf("✅ Team %s completed task %d: %s (+%d pts)", team, taskID, taskNames[taskID], taskPoints[taskID])

	broadcast <- task
}

func GetTeamScore(team string) models.TeamScore {
	progressLock.RLock()
	defer progressLock.RUnlock()

	score := models.TeamScore{
		Team:      team,
		Completed: []models.TaskProgress{},
	}

	for _, task := range progress[team] {
		score.TotalScore += task.Points
		score.Completed = append(score.Completed, task)
	}

	return score
}

func GetAllScores() []models.TeamScore {
	scores := []models.TeamScore{}
	for _, team := range middleware.Teams {
		scores = append(scores, GetTeamScore(team))
	}
	return scores
}

func ResetTeam(team string) {
	progressLock.Lock()
	defer progressLock.Unlock()
	progress[team] = make(map[int]models.TaskProgress)

	// Clear from persistent database
	if err := resetTeamInDB(team); err != nil {
		log.Printf("⚠️  Failed to reset team in DB: %v", err)
	}

	log.Printf("🔄 Team %s progress reset", team)
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
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				client.Close()
				failed = append(failed, client)
			}
		}
		clientsLock.RUnlock()

		if len(failed) > 0 {
			clientsLock.Lock()
			for _, client := range failed {
				delete(clients, client)
			}
			clientsLock.Unlock()
		}
	}
}
