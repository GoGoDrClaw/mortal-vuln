package tracker

import (
	"database/sql"
	"log"
	"vulnnotes/models"

	_ "github.com/mattn/go-sqlite3"
)

var progressDB *sql.DB

func InitProgressDB() error {
	var err error
	progressDB, err = sql.Open("sqlite3", "/data/progress.db")
	if err != nil {
		return err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS completed_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team TEXT NOT NULL,
		task_id INTEGER NOT NULL,
		task_name TEXT NOT NULL,
		points INTEGER NOT NULL,
		timestamp TEXT NOT NULL,
		details TEXT,
		UNIQUE(team, task_id)
	);
	CREATE INDEX IF NOT EXISTS idx_team ON completed_tasks(team);
	`

	_, err = progressDB.Exec(createTable)
	if err != nil {
		return err
	}

	log.Println("✅ Progress database initialized: data/progress.db")
	return loadProgressFromDB()
}

func loadProgressFromDB() error {
	rows, err := progressDB.Query(`
		SELECT team, task_id, task_name, points, timestamp, details
		FROM completed_tasks
		ORDER BY id ASC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	progressLock.Lock()
	defer progressLock.Unlock()

	count := 0
	for rows.Next() {
		var task models.TaskProgress
		err := rows.Scan(&task.Team, &task.TaskID, &task.TaskName, &task.Points, &task.Timestamp, &task.Details)
		if err != nil {
			log.Printf("⚠️  Error loading task: %v", err)
			continue
		}
		progress[task.Team][task.TaskID] = task
		count++
	}

	if count > 0 {
		log.Printf("📊 Loaded %d completed tasks from database", count)
	}
	return nil
}

func saveTaskToDB(task models.TaskProgress) error {
	_, err := progressDB.Exec(`
		INSERT OR IGNORE INTO completed_tasks (team, task_id, task_name, points, timestamp, details)
		VALUES (?, ?, ?, ?, ?, ?)
	`, task.Team, task.TaskID, task.TaskName, task.Points, task.Timestamp, task.Details)
	return err
}

func resetTeamInDB(team string) error {
	_, err := progressDB.Exec(`DELETE FROM completed_tasks WHERE team = ?`, team)
	return err
}

