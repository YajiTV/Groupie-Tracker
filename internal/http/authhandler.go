package httphandlers

import (
	"log"
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/storage"
	"github.com/YajiTV/groupie-tracker/internal/templates"
)

// LoginPageHandler affiche la page de connexion
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// Si déjà connecté, rediriger vers le profil
	if auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	data := struct {
		Title   string
		Error   string
		Success string
	}{
		Title:   "Connexion",
		Error:   r.URL.Query().Get("error"),
		Success: r.URL.Query().Get("success"),
	}

	templates.Templates.ExecuteTemplate(w, "login.gohtml", data)
}

// LoginHandler gère la connexion
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Utiliser la logique métier
	sessionID, err := authenticateUser(username, password)
	if err != nil {
		log.Printf("Login error (user=%q): %v", username, err)
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	// Créer le cookie de session
	auth.SetCookie(w, sessionID)
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// RegisterPageHandler affiche la page d'inscription
func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	// Si déjà connecté, rediriger vers le profil
	if auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	data := struct {
		Title string
		Error string
	}{
		Title: "Inscription",
		Error: r.URL.Query().Get("error"),
	}

	templates.Templates.ExecuteTemplate(w, "register.gohtml", data)
}

// RegisterHandler gère l'inscription
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Utiliser la logique métier
	err := registerNewUser(username, email, password)
	if err != nil {
		// IMPORTANT: afficher l'erreur réelle dans les logs
		log.Printf("Register error (user=%q, email=%q): %v", username, email, err)

		errorCode := getErrorCode(err)
		http.Redirect(w, r, "/register?error="+errorCode, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login?success=registered", http.StatusSeeOther)
}

// LogoutHandler gère la déconnexion
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err == nil {
		auth.Store.DeleteSession(cookie.Value)
	}

	auth.ClearCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ProfileHandler affiche le profil de l'utilisateur
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier l'authentification
	session, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Récupérer l'utilisateur complet
	user, err := storage.GetUserByID(session.UserID)
	if err != nil {
		http.Error(w, "Utilisateur introuvable", http.StatusNotFound)
		return
	}

	data := struct {
		Title   string
		User    interface{}
		Success string
		Error   string
	}{
		Title:   "Mon profil",
		User:    user,
		Success: r.URL.Query().Get("success"),
		Error:   r.URL.Query().Get("error"),
	}

	templates.Templates.ExecuteTemplate(w, "profile.gohtml", data)
}

// UpdateProfileHandler met à jour le profil de l'utilisateur
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Vérifier l'authentification
	session, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	bio := r.FormValue("bio")

	// Utiliser la logique métier
	err := updateUserProfile(session.UserID, bio)
	if err != nil {
		log.Printf("UpdateProfile error (userID=%d): %v", session.UserID, err)
		http.Redirect(w, r, "/profile?error=update", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile?success=updated", http.StatusSeeOther)
}
