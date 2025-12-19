package app

import (
	"log"
	"net/http"

	httphandlers "github.com/YajiTV/groupie-tracker/internal/http"
	"github.com/YajiTV/groupie-tracker/internal/storage"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Pages publiques
	mux.HandleFunc("/", httphandlers.HomeHandler)
	mux.HandleFunc("/artist/", httphandlers.ArtistHandler)
	mux.HandleFunc("/search", httphandlers.SearchHandler)
	mux.HandleFunc("/filters", httphandlers.FiltersHandler)
	mux.HandleFunc("/api/suggestions", httphandlers.SuggestionsHandler)

	// Authentification
	mux.HandleFunc("/login", httphandlers.LoginPageHandler)
	mux.HandleFunc("/register", httphandlers.RegisterPageHandler)
	mux.HandleFunc("/logout", httphandlers.LogoutHandler)
	mux.HandleFunc("/auth/login", httphandlers.LoginHandler)
	mux.HandleFunc("/auth/register", httphandlers.RegisterHandler)

	// Pages protégées
	mux.HandleFunc("/profile", httphandlers.ProfileHandler)
	mux.HandleFunc("/profile/update", httphandlers.UpdateProfileHandler)

	return mux
}

func Start() {
	// Initialiser le stockage
	if err := storage.InitUsers(); err != nil {
		log.Fatalf("Erreur initialisation stockage: %v", err)
	}

	log.Printf("Serveur sur http://localhost%s\n", Port)
	log.Fatal(http.ListenAndServe(Port, SetupRouter()))
}
