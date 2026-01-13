package httphandlers

import (
	"errors"
	"log"
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
)

// authenticateUser authentifie un utilisateur et retourne un sessionID
func authenticateUser(username, password string) (string, error) {
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
		return ErrUserExists
	}

	log.Printf("Nouvel utilisateur créé: %s", username)
	return nil
}

// validateRegistrationFields valide les données d'inscription
func validateRegistrationFields(username, email, password string) error {
	// Vérifier les champs vides
	if username == "" || email == "" || password == "" {
		return ErrEmptyFields
	}

	// Vérifier la longueur du mot de passe
	if len(password) < 6 {
		return ErrPasswordTooShort
	}

	// Vérifier la validité de l'email (basique)
	if !isValidEmail(email) {
		return errors.New("email invalide")
	}

	// Vérifier la longueur du username
	if len(username) < 3 {
		return errors.New("nom d'utilisateur trop court")
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
		return errors.New("bio trop longue")
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
	default:
		return "server"
	}
}

// isValidEmail vérifie si un email est valide (validation basique)
func isValidEmail(email string) bool {
	// Validation basique
	if len(email) < 3 {
		return false
	}

	// Vérifier la présence de @ et .
	atIndex := -1
	dotIndex := -1

	for i, char := range email {
		if char == '@' {
			atIndex = i
		}
		if char == '.' && atIndex != -1 {
			dotIndex = i
		}
	}

	// @ doit être présent, pas au début, et . doit être après @
	return atIndex > 0 && dotIndex > atIndex+1 && dotIndex < len(email)-1
}
