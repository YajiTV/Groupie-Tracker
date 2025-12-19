package httphandlers

import (
	"log"
	"net/http"
	"time"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/models"
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

	errorMsg := r.URL.Query().Get("error")
	successMsg := r.URL.Query().Get("success")

	data := struct {
		Title   string
		Error   string
		Success string
	}{
		Title:   "Connexion",
		Error:   errorMsg,
		Success: successMsg,
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

	// Récup l'utilisateur
	user, err := storage.GetUserByUsername(username)
	if err != nil {
		log.Println("Utilisateur introuvable:", username)
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	// check mpd
	if !auth.CheckPassword(user.Password, password) {
		log.Println("Mot de passe pas bon:", username)
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	// Créer une session
	sessionID := auth.Store.CreateSession(user.ID, user.Username)
	auth.SetCookie(w, sessionID)

	log.Printf("Connexion réussie: %s", username)
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// RegisterPageHandler affiche la page d'inscription
func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	// Si déjà connecté, rediriger vers le profil
	if auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	errorMsg := r.URL.Query().Get("error")

	data := struct {
		Title string
		Error string
	}{
		Title: "Inscription",
		Error: errorMsg,
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

	// Validation basique
	if username == "" || email == "" || password == "" {
		http.Redirect(w, r, "/register?error=empty", http.StatusSeeOther)
		return
	}

	if len(password) < 6 {
		http.Redirect(w, r, "/register?error=short", http.StatusSeeOther)
		return
	}

	// Hash du mot de passe
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Println("Erreur de hash:", err)
		http.Redirect(w, r, "/register?error=server", http.StatusSeeOther)
		return
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
		log.Println("Erreur création utilisateur:", err)
		http.Redirect(w, r, "/register?error=exists", http.StatusSeeOther)
		return
	}

	log.Printf("Nouvel utilisateur créé: %s", username)
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
		Title string
		User  interface{}
	}{
		Title: "Mon profil",
		User:  user,
	}

	templates.Templates.ExecuteTemplate(w, "profile.gohtml", data)
}
