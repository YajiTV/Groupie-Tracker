package models

import "time"

// User represents a user
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` // bcrypt hashed
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
}

// UserData contains all users (for JSON)
type UserData struct {
	Users  []User `json:"users"`
	LastID int    `json:"last_id"`
}
