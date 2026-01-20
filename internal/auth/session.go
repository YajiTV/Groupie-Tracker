package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SessionStore stocke les sessions en mémoire
type SessionStore struct {
	sessions map[string]*SessionData
	mutex    sync.RWMutex
}

// SessionData contient les données d'une session
type SessionData struct {
	UserID    int
	Username  string
	ExpiresAt time.Time
}

// Store global des sessions
var Store = &SessionStore{
	sessions: make(map[string]*SessionData),
}

const (
	SessionCookieName = "session_id"
	SessionDuration   = 24 * time.Hour
)

// GenerateSessionID génère un ID unique
func GenerateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateSession crée une nouvelle session
func (s *SessionStore) CreateSession(userID int, username string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sessionID := GenerateSessionID()
	s.sessions[sessionID] = &SessionData{
		UserID:    userID,
		Username:  username,
		ExpiresAt: time.Now().Add(SessionDuration),
	}

	return sessionID
}

// GetSession récupère une session
func (s *SessionStore) GetSession(sessionID string) (*SessionData, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Vérifier l'expiration
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, sessionID)
		return nil, false
	}

	return session, true
}

// DeleteSession supprime une session
func (s *SessionStore) DeleteSession(sessionID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.sessions, sessionID)
}

// SetCookie définit le cookie de session
func SetCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(SessionDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearCookie supprime le cookie de session
func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// GetUserFromRequest récupère l'utilisateur depuis la requête
func GetUserFromRequest(r *http.Request) (*SessionData, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, false
	}

	return Store.GetSession(cookie.Value)
}

// IsAuthenticated vérifie si l'utilisateur est connecté
func IsAuthenticated(r *http.Request) bool {
	_, ok := GetUserFromRequest(r)
	return ok
}

// HashPassword hash un mot de passe avec bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPassword vérifie un mot de passe
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
