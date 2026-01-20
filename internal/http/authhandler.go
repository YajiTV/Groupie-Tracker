package httphandlers

import (
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/storage"
	"github.com/YajiTV/groupie-tracker/internal/templates"
)

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	sessionID, err := authenticateUser(username, password)
	if err != nil {
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	auth.SetCookie(w, sessionID)
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	err := registerNewUser(username, email, password)
	if err != nil {
		errorCode := getErrorCode(err)
		http.Redirect(w, r, "/register?error="+errorCode, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login?success=registered", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err == nil {
		auth.Store.DeleteSession(cookie.Value)
	}

	auth.ClearCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	session, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	bio := r.FormValue("bio")

	err := updateUserProfile(session.UserID, bio)
	if err != nil {
		http.Redirect(w, r, "/profile?error=update", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile?success=updated", http.StatusSeeOther)
}
