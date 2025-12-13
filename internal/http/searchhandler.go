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
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	// Si pas de query, rediriger vers l'accueil
	if query == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Récupérer tous les artistes
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur API", 500)
		log.Println("Erreur lors de la récupération des artistes:", err)
		return
	}

	// Filtrer les artistes
	filteredArtists := searchArtists(allArtists, query)

	data := SearchData{
		Title:   "Résultats de recherche",
		Artists: filteredArtists,
		Query:   query,
		Count:   len(filteredArtists),
	}

	templates.Templates.ExecuteTemplate(w, "search.gohtml", data)
}

// searchArtists filtre les artistes selon la query (nom ou membres)
func searchArtists(artists []util.Artist, query string) []util.Artist {
	var results []util.Artist
	queryLower := strings.ToLower(query)

	for _, artist := range artists {
		// Recherche dans le nom de l'artiste (insensible à la casse)
		if strings.Contains(strings.ToLower(artist.Name), queryLower) {
			results = append(results, artist)
			continue
		}

		// Recherche dans les membres
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), queryLower) {
				results = append(results, artist)
				break // Éviter les doublons
			}
		}
	}

	return results
}
