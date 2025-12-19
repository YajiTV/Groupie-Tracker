package models

import "time"

// User représente un utilisateur
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` //  bcrypt ydays
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
}

// UserData contient tous les utilisateurs (pour JSON)
type UserData struct {
	Users  []User `json:"users"`
	LastID int    `json:"last_id"`
}
