package httphandlers

import (
	"fmt"
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

// HomeHandler gère la page d'accueil avec filtres
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier que c'est bien la route racine
	if r.URL.Path != "/" {
		NotFoundHandler(w, r)
		return
	}

	// Parser les filtres depuis l'URL
	filters := parseHomeFilters(r)

	// Récupérer tous les artistes
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		return
	}

	// Récupérer les relations (locations) pour tous les artistes
	artistLocations := fetchArtistLocations()

	// Récupérer tous les lieux disponibles pour le filtre
	allLocations := getAllUniqueLocationsFromRelations(artistLocations)

	// Appliquer les filtres
	displayedArtists := applyHomeFilters(allArtists, filters, artistLocations)

	// Préparer les données pour le template
	data := struct {
		Title           string
		Artists         []util.Artist
		Filters         HomeFilters
		AllLocations    []string
		IsAuthenticated bool
	}{
		Title:           "Groupie Tracker",
		Artists:         displayedArtists,
		Filters:         filters,
		AllLocations:    allLocations,
		IsAuthenticated: auth.IsAuthenticated(r),
	}

	// Rendre le template
	if err := templates.Templates.ExecuteTemplate(w, "home.gohtml", data); err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors du rendu du template: %v", err), http.StatusInternalServerError)
		return
	}
}
