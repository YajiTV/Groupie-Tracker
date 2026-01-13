package httphandlers

import (
	"log"
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

// HomeFilters représente les filtres de la page d'accueil
type HomeFilters struct {
	CreationYearMin int
	CreationYearMax int
	AlbumYearMin    int
	AlbumYearMax    int
	MemberCounts    []int
	Locations       []string
	Query           string
}

// HomeHandler gère la page d'accueil avec filtres
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🏠 HomeHandler appelé avec URL: %s", r.URL.String())

	// Vérifier que c'est bien la route racine
	if r.URL.Path != "/" {
		NotFoundHandler(w, r)
		return
	}

	// Parser les filtres depuis l'URL
	filters := parseHomeFilters(r)
	log.Printf("🔍 Filtres détectés: Creation(%d-%d), Albums(%d-%d), Membres%v, Lieux%v, Query:'%s'",
		filters.CreationYearMin, filters.CreationYearMax,
		filters.AlbumYearMin, filters.AlbumYearMax,
		filters.MemberCounts, filters.Locations, filters.Query)

	// Récupérer tous les artistes
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		log.Printf("Erreur API: %v", err)
		return
	}
	log.Printf("📊 %d artistes récupérés depuis l'API", len(allArtists))

	// Récupérer tous les lieux disponibles pour le filtre
	allLocations := getAllUniqueLocations()

	// Appliquer les filtres
	displayedArtists := applyHomeFilters(allArtists, filters)
	log.Printf("✅ %d artistes envoyés au template", len(displayedArtists))

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
	panic("unimplemented")
}
