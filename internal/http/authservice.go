package httphandlers

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/models"
	"github.com/YajiTV/groupie-tracker/internal/storage"
)

// Erreurs personnalisées
var (
	ErrEmptyFields        = errors.New("champs vides")
	ErrPasswordTooShort   = errors.New("mot de passe trop court")
	ErrUserExists         = errors.New("utilisateur existe déjà")
	ErrInvalidCredentials = errors.New("identifiants invalides")
	ErrUserNotFound       = errors.New("utilisateur introuvable")
	ErrServerError        = errors.New("erreur serveur")

	ErrInvalidEmail     = errors.New("email invalide")
	ErrUsernameTooShort = errors.New("nom d'utilisateur trop court")
	ErrBioTooLong       = errors.New("bio trop longue")
)

// authenticateUser authentifie un utilisateur et retourne un sessionID
func authenticateUser(username, password string) (string, error) {
	username = strings.TrimSpace(username)

	user, err := storage.GetUserByUsername(username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !auth.CheckPassword(user.Password, password) {
		return "", ErrInvalidCredentials
	}

	sessionID := auth.Store.CreateSession(user.ID, user.Username)
	return sessionID, nil
}

// registerNewUser crée un nouvel utilisateur
func registerNewUser(username, email, password string) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if err := validateRegistrationFields(username, email, password); err != nil {
		return err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return ErrServerError
	}

	user := models.User{
		Username:  username,
		Email:     email,
		Password:  hash,
		AvatarURL: "/static/img/logo.png",
		Bio:       "",
		CreatedAt: time.Now(),
	}

	_, err = storage.CreateUser(user)
	if err != nil {
		return ErrUserExists
	}

	return nil
}

// validateRegistrationFields valide les données d'inscription
func validateRegistrationFields(username, email, password string) error {
	if username == "" || email == "" || password == "" {
		return ErrEmptyFields
	}

	if len(password) < 6 {
		return ErrPasswordTooShort
	}

	if len(username) < 3 {
		return ErrUsernameTooShort
	}

	if !isValidEmail(email) {
		return ErrInvalidEmail
	}

	return nil
}

// updateUserProfile met à jour le profil d'un utilisateur
func updateUserProfile(userID int, bio string) error {
	user, err := storage.GetUserByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	if len(bio) > 500 {
		return ErrBioTooLong
	}

	user.Bio = bio

	if err := storage.UpdateUser(*user); err != nil {
		return ErrServerError
	}

	return nil
}

// getErrorCode convertit une erreur en code d'erreur pour l'URL
func getErrorCode(err error) string {
	switch err {
	case ErrEmptyFields:
		return "empty"
	case ErrPasswordTooShort:
		return "short"
	case ErrUserExists:
		return "exists"
	case ErrInvalidCredentials:
		return "invalid"
	case ErrInvalidEmail:
		return "email"
	case ErrUsernameTooShort:
		return "username"
	case ErrBioTooLong:
		return "bio"
	default:
		return "server"
	}
}

// isValidEmail vérifie si un email est valide
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
