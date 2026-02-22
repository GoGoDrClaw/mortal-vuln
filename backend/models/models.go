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

// Claims is intentionally insecure (CTF: password exposed in JWT payload).
type Claims struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Password  string `json:"password"`
	SessionID string `json:"sessionId"`
	jwt.RegisteredClaims
}
