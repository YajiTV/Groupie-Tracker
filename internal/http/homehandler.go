package httphandlers

import (
	"fmt"
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

// HomeHandler handles the home page with filters
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Check that this is the root route
	if r.URL.Path != "/" {
		NotFoundHandler(w, r)
		return
	}

	// Parse filters from URL
	filters := parseHomeFilters(r)

	// Retrieve all artists
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		return
	}

	// Retrieve relations (locations) for all artists
	artistLocations := fetchArtistLocations()

	// Retrieve all available locations for the filter
	allLocations := getAllUniqueLocationsFromRelations(artistLocations)

	// Apply filters
	displayedArtists := applyHomeFilters(allArtists, filters, artistLocations)

	// Prepare data for the template
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

	// Render the template
	if err := templates.Templates.ExecuteTemplate(w, "home.gohtml", data); err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors du rendu du template: %v", err), http.StatusInternalServerError)
		return
	}
}
