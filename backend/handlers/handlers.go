package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"vulnnotes/db"
	"vulnnotes/middleware"
	"vulnnotes/models"
	"vulnnotes/tracker"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
)

var validCharacters = map[string]bool{
	"scorpion": true, "subzero": true, "liukang": true, "kitana": true,
	"raiden": true, "jax": true, "mileena": true, "kunglao": true,
	"sonya": true, "shangtsung": true, "kano": true, "nightwolf": true,
	"cyrax": true, "sektor": true, "kabal": true, "jade": true,
	"sindel": true, "ermac": true, "sheeva": true, "stryker": true,
	"smoke": true, "noobsaibot": true, "motaro": true, "shaokahn": true,
}

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func ReadBody(r *http.Request) map[string]string {
	m := map[string]string{}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		json.NewDecoder(r.Body).Decode(&m)
	} else {
		r.ParseForm()
		for k, v := range r.PostForm {
			if len(v) > 0 {
				m[k] = v[0]
			}
		}
	}
	return m
}

func currentUser(r *http.Request) (int, string, string) {
	id, _ := strconv.Atoi(r.Header.Get("X-User-ID"))
	return id, r.Header.Get("X-Username"), r.Header.Get("X-Role")
}

func sessionDB(r *http.Request) *sql.DB {
	return middleware.GetSessionDB(r.Header.Get("X-Session-ID"))
}

// ── Session management ────────────────────────────────────────────

// InitSessions ensures /data/sessions directory exists.
func InitSessions() {
	os.MkdirAll("/data/sessions", 0755)
	log.Println("✅ Sessions storage initialised: /data/sessions/")
}

// POST /api/session/new — create a new player session.
func NewSession(w http.ResponseWriter, r *http.Request) {
	body := ReadBody(r)
	nickname := strings.TrimSpace(body["nickname"])
	character := strings.ToLower(strings.TrimSpace(body["character"]))

	if nickname == "" || len(nickname) > 50 {
		WriteJSON(w, 400, map[string]string{"error": "Nickname must be 1–50 characters"})
		return
	}
	if !validCharacters[character] {
		WriteJSON(w, 400, map[string]string{"error": "Invalid character"})
		return
	}

	sess, err := db.CreateSession(nickname, character)
	if err != nil {
		if err == db.ErrDuplicateNicknameChar {
			WriteJSON(w, 409, map[string]string{"error": "This name is already taken for this character. Choose a different character or use another name."})
			return
		}
		log.Printf("❌ CreateSession: %v", err)
		WriteJSON(w, 500, map[string]string{"error": "Failed to create session"})
		return
	}

	// Seed isolated CTF database for this session
	ctfDB := middleware.GetSessionDB(sess.ID)
	seedDB(ctfDB)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sess.ID,
		Path:     "/",
		HttpOnly: false, // intentionally readable by JS (CTF)
		MaxAge:   30 * 24 * 60 * 60,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})

	go tracker.BroadcastNewSession()

	WriteJSON(w, 200, map[string]string{
		"saveCode":  sess.DisplayCode(),
		"sessionId": sess.ID,
		"nickname":  sess.Nickname,
		"character": sess.Character,
	})
}

// POST /api/session/restore — restore a session by save code.
func RestoreSession(w http.ResponseWriter, r *http.Request) {
	body := ReadBody(r)
	code := strings.TrimSpace(body["saveCode"])

	sess, err := db.GetSessionByCode(code)
	if err != nil {
		log.Printf("❌ GetSessionByCode: %v", err)
		WriteJSON(w, 500, map[string]string{"error": "Database error"})
		return
	}
	if sess == nil {
		WriteJSON(w, 404, map[string]string{"error": "Invalid save code"})
		return
	}

	// Re-seed CTF database (fresh environment on restore)
	ctfDB := middleware.GetSessionDB(sess.ID)
	seedDB(ctfDB)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sess.ID,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   30 * 24 * 60 * 60,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})

	score := tracker.GetSessionScore(sess.ID)
	WriteJSON(w, 200, map[string]any{
		"saveCode":       sess.DisplayCode(),
		"sessionId":      sess.ID,
		"nickname":       sess.Nickname,
		"character":      sess.Character,
		"message":        "Session restored",
		"tasksCompleted": len(score.Completed),
		"totalScore":     score.TotalScore,
	})
}

// GET /api/session/check — return current session info from cookie.
func CheckSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_id")
	if err != nil || c.Value == "" {
		WriteJSON(w, 401, map[string]string{"error": "No session"})
		return
	}

	sess, err := db.GetSessionByID(c.Value)
	if err != nil || sess == nil {
		WriteJSON(w, 401, map[string]string{"error": "Invalid session"})
		return
	}

	WriteJSON(w, 200, map[string]string{
		"saveCode":  sess.DisplayCode(),
		"sessionId": sess.ID,
		"nickname":  sess.Nickname,
		"character": sess.Character,
	})
}

// GET /api/session/progress — completed tasks for current session.
func SessionProgress(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_id")
	if err != nil || c.Value == "" {
		WriteJSON(w, 401, map[string]string{"error": "No session"})
		return
	}
	sess, err := db.GetSessionByID(c.Value)
	if err != nil || sess == nil {
		WriteJSON(w, 401, map[string]string{"error": "Invalid session"})
		return
	}
	score := tracker.GetSessionScore(sess.ID)
	WriteJSON(w, 200, map[string]interface{}{
		"totalScore": score.TotalScore,
		"completed":  score.Completed,
	})
}

// GET /api/characters — list all available characters.
func GetCharacters(w http.ResponseWriter, r *http.Request) {
	chars := make([]string, 0, len(validCharacters))
	for k := range validCharacters {
		chars = append(chars, k)
	}
	WriteJSON(w, 200, chars)
}

// ── CTF handlers ──────────────────────────────────────────────────

func Login(w http.ResponseWriter, r *http.Request) {
	d := sessionDB(r)
	body := ReadBody(r)
	username := body["username"]
	password := body["password"]
	sessionID := r.Header.Get("X-Session-ID")

	query := fmt.Sprintf(
		"SELECT id, username, password, role FROM users WHERE username='%s' AND password='%s'",
		username, password,
	)

	var user models.User
	err := d.QueryRow(query).Scan(&user.ID, &user.Username, &user.Password, &user.Role)

	if err == nil {
		switch user.Username {
		case "alice":
			if password == "123456" {
				tracker.CompleteTask(sessionID, 0, "Brute forced alice's password")
			}
		case "admin":
			if password != "Sup3r_S3cret_Adm1n!" {
				tracker.CompleteTask(sessionID, 4, fmt.Sprintf("SQL Injection: username='%s'", username))
			} else {
				tracker.CompleteTask(sessionID, 5, "Logged in as admin with correct password")
			}
		}
	}

	if err != nil {
		WriteJSON(w, 401, map[string]string{
			"error": "Invalid credentials",
			"debug": err.Error(),
			"query": query,
			"stack": string(debug.Stack()),
		})
		return
	}

	claims := models.Claims{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Password:  user.Password,
		SessionID: sessionID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(middleware.JWTSecret)

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenStr,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   30 * 24 * 60 * 60,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})

	WriteJSON(w, 200, map[string]string{"message": "Login successful", "token": tokenStr})
}

func GetNotes(w http.ResponseWriter, r *http.Request) {
	d := sessionDB(r)
	uid, _, _ := currentUser(r)
	rows, err := d.Query("SELECT id, user_id, title, content, created FROM notes WHERE user_id = ?", uid)
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	notes := []models.Note{}
	for rows.Next() {
		var n models.Note
		rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Created)
		notes = append(notes, n)
	}
	WriteJSON(w, 200, notes)
}

func GetNote(w http.ResponseWriter, r *http.Request) {
	d := sessionDB(r)
	uid, _, _ := currentUser(r)
	sessionID := r.Header.Get("X-Session-ID")
	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]

	var n models.Note
	err := d.QueryRow(
		"SELECT id, user_id, title, content, created FROM notes WHERE id = ?", idStr,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Created)

	if err != nil {
		WriteJSON(w, 404, map[string]string{"error": "Not found"})
		return
	}

	if n.UserID != uid && n.UserID == 1 {
		tracker.CompleteTask(sessionID, 3, fmt.Sprintf("Read admin note ID=%d", n.ID))
	}

	WriteJSON(w, 200, n)
}

func CreateNote(w http.ResponseWriter, r *http.Request) {
	d := sessionDB(r)
	uid, _, _ := currentUser(r)
	sessionID := r.Header.Get("X-Session-ID")
	body := ReadBody(r)

	title := body["title"]
	content := body["content"]

	if strings.Contains(content, "<script") || strings.Contains(content, "onerror") {
		tracker.CompleteTask(sessionID, 6, "XSS payload detected")
	}

	if strings.ToUpper(title) == "FINISH HIM" && strings.Contains(content, "↑↑↓↓←→←→BA") {
		tracker.CompleteTask(sessionID, 10, "🎮 FATALITY! Mortal Kombat easter egg discovered!")
	}

	res, err := d.Exec("INSERT INTO notes (user_id, title, content) VALUES (?, ?, ?)", uid, title, content)
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	WriteJSON(w, 200, map[string]int64{"id": id})
}

func DeleteNote(w http.ResponseWriter, r *http.Request) {
	d := sessionDB(r)
	uid, _, _ := currentUser(r)
	sessionID := r.Header.Get("X-Session-ID")
	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]

	var ownerID int
	d.QueryRow("SELECT user_id FROM notes WHERE id = ?", idStr).Scan(&ownerID)

	origin := r.Header.Get("Origin")
	if origin != "" {
		host := r.Host
		if i := strings.LastIndex(host, ":"); i != -1 {
			host = host[:i]
		}
		baseDomain := strings.TrimPrefix(host, "api.")
		originHost := origin
		if i := strings.Index(originHost, "://"); i != -1 {
			originHost = originHost[i+3:]
		}
		if i := strings.LastIndex(originHost, ":"); i != -1 {
			originHost = originHost[:i]
		}
		isLegit := originHost == baseDomain ||
			originHost == "dashboard."+baseDomain ||
			originHost == "www."+baseDomain
		if !isLegit {
			tracker.CompleteTask(sessionID, 9, fmt.Sprintf("CSRF: deleted note ID=%s from origin=%s", idStr, origin))
		}
	}

	if ownerID != uid && ownerID != 0 {
		tracker.CompleteTask(sessionID, 7, fmt.Sprintf("Deleted note ID=%s owned by user %d", idStr, ownerID))
	}

	d.Exec("DELETE FROM notes WHERE id = ?", idStr)
	WriteJSON(w, 200, map[string]string{"message": "Deleted"})
}

func AdminUsers(w http.ResponseWriter, r *http.Request) {
	d := sessionDB(r)
	_, _, role := currentUser(r)

	if role != "admin" {
		WriteJSON(w, 403, map[string]string{"error": "Admin only"})
		return
	}

	rows, _ := d.Query("SELECT id, username, password, role FROM users")
	defer rows.Close()
	users := []models.User{}
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Password, &u.Role)
		users = append(users, u)
	}
	WriteJSON(w, 200, users)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	d := sessionDB(r)
	uid, _, _ := currentUser(r)
	body := ReadBody(r)
	d.Exec("UPDATE users SET password = ? WHERE id = ?", body["newPassword"], uid)
	WriteJSON(w, 200, map[string]string{"message": "Password changed"})
}

func SubmitFlag(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")

	buf := make([]byte, 256)
	n, _ := r.Body.Read(buf)
	flag := strings.TrimSpace(string(buf[:n]))

	validFlags := map[string]int{
		"password":            1,
		"exp":                 2,
		"Sup3r_S3cret_Adm1n!": 5,
	}

	taskID, ok := validFlags[flag]
	if !ok {
		WriteJSON(w, 400, map[string]string{"error": "Wrong flag"})
		return
	}

	score := tracker.GetSessionScore(sessionID)
	for _, completed := range score.Completed {
		if completed.TaskID == taskID {
			WriteJSON(w, 400, map[string]string{"error": "Task already completed"})
			return
		}
	}

	tracker.CompleteTask(sessionID, taskID, "")
	WriteJSON(w, 200, map[string]string{"message": "Correct!"})
}

func Reset(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_id")
	if err != nil || c.Value == "" {
		WriteJSON(w, 400, map[string]string{"error": "No session"})
		return
	}
	d := middleware.GetSessionDB(c.Value)
	if d == nil {
		WriteJSON(w, 400, map[string]string{"error": "Session not found"})
		return
	}
	seedDB(d)
	WriteJSON(w, 200, map[string]string{"message": "Database reset"})
}

func GetScores(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, 200, tracker.GetAllScores())
}

func seedDB(d *sql.DB) {
	if d == nil {
		return
	}
	d.Exec(`DROP TABLE IF EXISTS notes`)
	d.Exec(`DROP TABLE IF EXISTS users`)
	d.Exec(`
	CREATE TABLE users (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role     TEXT DEFAULT 'user'
	);
	CREATE TABLE notes (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id  INTEGER REFERENCES users(id),
		title    TEXT,
		content  TEXT,
		created  DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	INSERT INTO users (username, password, role) VALUES
		('admin', 'Sup3r_S3cret_Adm1n!', 'admin'),
		('alice', '123456',                'user'),
		('bob',   'qwerty',               'user');
	INSERT INTO notes (user_id, title, content) VALUES
		(1, '⚔️ Imperial Decree', 'In the name of Shao Kahn, Ruler of Outworld:

All fighters must report to the Coliseum at dawn. Failure to appear is punishable by death. Or worse.'),
		(1, '🌐 Outworld Servers', 'prod-outworld.shao-kahn.local:5432
user=root
pass=t0p_s3cret

⚠ TOP SECRET ⚠
DO NOT SHARE WITH EARTHREALM WARRIORS
Shang Tsung is personally responsible for security'),
		(2, 'How to submit a flag', 'curl -X POST https://api.test-your-might.ru/api/flag \
  --cookie "session_id=<your_session_id>" \
  -d "your_flag"'),
		(2, 'Easy tasks (1 point each)', 'Task 0 — Log in to the app as alice
Task 1 — Find critical information in the JWT token
  Required: submit the field name via /api/flag
Task 2 — Something important is missing from the JWT token
  Required: submit the field name via /api/flag
Task 3 — Gain access to someone else''s notes'),
		(2, 'Medium tasks (2 points each)', 'Task 4 — Log in as administrator
Task 5 — Find out the administrator password and log in using it
Task 6 — Perform an XSS attack
Task 7 — Delete someone else''s note'),
		(2, 'Hard tasks (3 points each)', 'Task 8 — Become admin by forging a JWT token
Task 9 — Perform an action on behalf of another user'),
		(3, '📝 TODO Before Mortal Kombat', 'Finish the security assignment before the tournament finals
Or Shao Kahn will expel you from the program');
	`)
}

func DetectNoneAlg(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	tokenStr := ""
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		tokenStr = strings.TrimPrefix(auth, "Bearer ")
	}

	parts := strings.Split(tokenStr, ".")
	if len(parts) == 3 && parts[2] == "" {
		tracker.CompleteTask(sessionID, 8, "JWT with alg:none detected")
	}
}

func TrackAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		DetectNoneAlg(w, r)
		next(w, r)
	}
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Secret-Hint", "54657374596F75724D69676874")
	w.Header().Set("X-Powered-By", "Kombat-Engine-v1.0")
	WriteJSON(w, 200, map[string]string{
		"status": "ok",
		"time":   time.Now().Format("15:04:05"),
	})
}

func Toasty(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		WriteJSON(w, 400, map[string]string{"error": "No session"})
		return
	}

	userAgent := r.Header.Get("User-Agent")
	isKombatWarrior := strings.Contains(strings.ToLower(userAgent), "scorpion") ||
		strings.Contains(strings.ToLower(userAgent), "sub-zero") ||
		strings.Contains(strings.ToLower(userAgent), "raiden")

	d := sessionDB(r)
	var count int
	d.QueryRow("SELECT COUNT(*) FROM notes WHERE UPPER(title) = 'FINISH HIM' AND content LIKE '%↑↑↓↓←→←→BA%'").Scan(&count)

	if count > 0 {
		response := map[string]any{
			"message":     "🎮 TOASTY! 🎮",
			"flag":        "FLAG{test_your_might_flawless_victory}",
			"achievement": "Mortal Kombat Easter Egg",
			"bonus":       "You've unlocked the ancient secrets of kombat!",
			"hints": []string{
				"The code was hidden in plain sight...",
				"↑↑↓↓←→←→BA - A legendary sequence",
				"FINISH HIM - The ultimate command",
			},
			"scorpion_says": "GET OVER HERE!",
			"sub_zero_says": "The secret is frozen no more",
		}
		if isKombatWarrior {
			response["special_bonus"] = "You've chosen your fighter wisely!"
			response["hidden_flag"] = "FLAG{raiden_consulted_with_the_elder_gods}"
		}
		WriteJSON(w, 200, response)
	} else {
		hint := "You must prove yourself worthy in kombat first..."
		if isKombatWarrior {
			hint = "Even warriors of the realm need the right kombat credentials. Create the victory note first!"
		}
		WriteJSON(w, 403, map[string]string{"error": "Access Denied", "hint": hint})
	}
}
