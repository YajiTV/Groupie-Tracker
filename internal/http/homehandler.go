package httphandlers

import (
	"log"
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

// HomeHandler gère la page d'accueil avec filtres
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier que c'est bien la route racine
	if r.URL.Path != "/" {
		Handler404(w, r)
		return
	}

	// Récupérer tous les artistes
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		log.Printf("Erreur API: %v", err)
		return
	}

	// Parser les filtres depuis l'URL
	filters := parseFilters(r)

	// Appliquer les filtres si présents
	var displayedArtists []util.Artist
	if hasActiveFilters(filters) {
		relationResponse, _ := util.FetchRelations()
		displayedArtists = applyFilters(allArtists, filters, relationResponse)
		log.Printf("🔍 Filtres appliqués: %d artistes trouvés", len(displayedArtists))
	} else {
		displayedArtists = allArtists
	}

	// Préparer les données pour le template
	data := struct {
		Title           string
		Artists         []util.Artist
		IsAuthenticated bool
	}{
		Title:           "Groupie Tracker",
		Artists:         displayedArtists,
		IsAuthenticated: auth.IsAuthenticated(r),
	}

	// Rendre le template
	if err := templates.Templates.ExecuteTemplate(w, "home.gohtml", data); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
		log.Printf("Erreur template: %v", err)
	}
}

func Handler404(w http.ResponseWriter, r *http.Request) {
	// TODO: implémenter la page 404
	http.Error(w, "404 - Page non trouvée", http.StatusNotFound)
}
