package httphandlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

type SearchData struct {
	Title   string
	Artists []util.Artist
	Query   string
	Count   int
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	// If no query, redirect to home
	if query == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Retrieve all artists
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur API", 500)
		log.Println("Error retrieving artists:", err)
		return
	}

	// Filter artists
	filteredArtists := searchArtists(allArtists, query)

	data := SearchData{
		Title:   "Résultats de recherche",
		Artists: filteredArtists,
		Query:   query,
		Count:   len(filteredArtists),
	}

	templates.Templates.ExecuteTemplate(w, "search.gohtml", data)
}

// searchArtists filters artists by query (name or members)
func searchArtists(artists []util.Artist, query string) []util.Artist {
	var results []util.Artist
	queryLower := strings.ToLower(query)

	for _, artist := range artists {
		// Search in artist name (case-insensitive)
		if strings.Contains(strings.ToLower(artist.Name), queryLower) {
			results = append(results, artist)
			continue
		}

		// Search in members
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), queryLower) {
				results = append(results, artist)
				break // Avoid duplicates
			}
		}
	}

	return results
}
