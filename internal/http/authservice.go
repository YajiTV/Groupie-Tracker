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

// Custom errors
var (
	ErrEmptyFields        = errors.New("empty fields")
	ErrPasswordTooShort   = errors.New("password too short")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrServerError        = errors.New("server error")

	ErrInvalidEmail     = errors.New("invalid email")
	ErrUsernameTooShort = errors.New("username too short")
	ErrBioTooLong       = errors.New("bio too long")
)

// authenticateUser authenticates a user and returns a sessionID
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

// registerNewUser creates a new user
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
		AvatarURL: "~/static/img/",
		Bio:       "",
		CreatedAt: time.Now(),
	}

	_, err = storage.CreateUser(user)
	if err != nil {
		return ErrUserExists
	}

	return nil
}

// validateRegistrationFields validates registration data
func validateRegistrationFields(username, email, password string) error {
	if username == "" || email == "" || password == "" {
		return ErrEmptyFields
	}

	if len(password) < 6 { // Recommended secure minimum
		return ErrPasswordTooShort
	}

	if len(username) < 3 { // Minimum to avoid too short usernames
		return ErrUsernameTooShort
	}

	if !isValidEmail(email) {
		return ErrInvalidEmail
	}

	return nil
}

// updateUserProfile updates a user's profile
func updateUserProfile(userID int, bio string) error {
	user, err := storage.GetUserByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	if len(bio) > 500 { // Limit to avoid too long bios
		return ErrBioTooLong
	}

	user.Bio = bio

	if err := storage.UpdateUser(*user); err != nil {
		return ErrServerError
	}

	return nil
}

// getErrorCode converts an error to an error code for the URL
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

// isValidEmail checks if an email is valid
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
