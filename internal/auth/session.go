package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SessionStore stores sessions in memory
type SessionStore struct {
	sessions map[string]*SessionData
	mutex    sync.RWMutex
}

// SessionData contains session data
type SessionData struct {
	UserID    int
	Username  string
	ExpiresAt time.Time
}

// Store is the global session store
var Store = &SessionStore{
	sessions: make(map[string]*SessionData),
}

const (
	SessionCookieName = "session_id"
	SessionDuration   = 24 * time.Hour // Session validity duration
)

// GenerateSessionID generates a unique 32-character hexadecimal session ID
func GenerateSessionID() string {
	b := make([]byte, 16) // 16 bytes = 32 hex characters
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateSession creates a new session
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

// GetSession retrieves a session
func (s *SessionStore) GetSession(sessionID string) (*SessionData, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, sessionID)
		return nil, false
	}

	return session, true
}

// DeleteSession deletes a session
func (s *SessionStore) DeleteSession(sessionID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.sessions, sessionID)
}

// SetCookie sets the session cookie
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

// ClearCookie removes the session cookie
func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// GetUserFromRequest retrieves the user from the request
func GetUserFromRequest(r *http.Request) (*SessionData, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, false
	}

	return Store.GetSession(cookie.Value)
}

// IsAuthenticated checks if the user is logged in
func IsAuthenticated(r *http.Request) bool {
	_, ok := GetUserFromRequest(r)
	return ok
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPassword verifies a password
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
