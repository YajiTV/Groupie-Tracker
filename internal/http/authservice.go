package httphandlers

import (
	"errors"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/models"
	"github.com/YajiTV/groupie-tracker/internal/storage"
)

// Erreurs personnalisées pour une meilleure gestion
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

	// Récupérer l'utilisateur
	user, err := storage.GetUserByUsername(username)
	if err != nil {
		log.Printf("Utilisateur introuvable: %s", username)
		return "", ErrInvalidCredentials
	}

	// Vérifier le mot de passe
	if !auth.CheckPassword(user.Password, password) {
		log.Printf("Mot de passe incorrect pour: %s", username)
		return "", ErrInvalidCredentials
	}

	// Créer une session
	sessionID := auth.Store.CreateSession(user.ID, user.Username)
	log.Printf("Connexion réussie: %s", username)

	return sessionID, nil
}

// registerNewUser crée un nouvel utilisateur
func registerNewUser(username, email, password string) error {
	// Normaliser un minimum
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	// Validation des champs
	if err := validateRegistrationFields(username, email, password); err != nil {
		return err
	}

	// Hash du mot de passe
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("Erreur de hash: %v", err)
		return ErrServerError
	}

	// Créer l'utilisateur
	user := models.User{
		Username:  username,
		Email:     email,
		Password:  hash,
		AvatarURL: "/static/img/default-avatar.png",
		Bio:       "",
		CreatedAt: time.Now(),
	}

	_, err = storage.CreateUser(user)
	if err != nil {
		log.Printf("Erreur création utilisateur: %v", err)
		// Si votre storage renvoie une erreur spécifique, vous pourrez l'affiner.
		return ErrUserExists
	}

	log.Printf("Nouvel utilisateur créé: %s", username)
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
	// Récupérer l'utilisateur
	user, err := storage.GetUserByID(userID)
	if err != nil {
		log.Printf("Utilisateur introuvable: %d", userID)
		return ErrUserNotFound
	}

	// Valider la bio
	if len(bio) > 500 {
		return ErrBioTooLong
	}

	// Mettre à jour la bio
	user.Bio = bio

	// Sauvegarder
	if err := storage.UpdateUser(*user); err != nil {
		log.Printf("Erreur mise à jour profil: %v", err)
		return ErrServerError
	}

	log.Printf("Profil mis à jour: %s", user.Username)
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

// isValidEmail vérifie si un email est valide (validation fiable)
func isValidEmail(email string) bool {
	// ParseAddress valide une adresse au format RFC (ex: "a@b.com") [web:13]
	_, err := mail.ParseAddress(email)
	return err == nil
}
