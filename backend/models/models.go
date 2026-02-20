package models

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Note struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Created string `json:"created"`
}

type Claims struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Password string `json:"password"`
	Team     string `json:"team"`
	jwt.RegisteredClaims
}

type TaskProgress struct {
	Team      string `json:"team"`
	TaskID    int    `json:"task_id"`
	TaskName  string `json:"task_name"`
	Points    int    `json:"points"`
	Timestamp string `json:"timestamp"`
	Details   string `json:"details,omitempty"`
}

type TeamScore struct {
	Team       string         `json:"team"`
	TotalScore int            `json:"total_score"`
	Completed  []TaskProgress `json:"completed"`
}
