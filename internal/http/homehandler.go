package httphandlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

type HomeData struct {
	Title   string
	Artists []util.Artist
}

// Structure pour les filtres (comme dans filtershandler.go)
type HomeFilters struct {
	CreationYearMin int
	CreationYearMax int
	AlbumYearMin    int
	AlbumYearMax    int
	MemberCounts    []int
	Locations       []string
	Query           string
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		NotFoundHandler(w, r)
		return
	}

	log.Printf("🏠 ÉTAPE 2: HomeHandler appelé - URL: %s", r.URL.RawQuery)

	// Parser les filtres depuis l'URL
	filters := parseHomeFilters(r)
	log.Printf("🔍 ÉTAPE 2: Filtres détectés: Creation(%d-%d), Album(%d-%d), Members%v, Locations%v, Query:'%s'",
		filters.CreationYearMin, filters.CreationYearMax,
		filters.AlbumYearMin, filters.AlbumYearMax,
		filters.MemberCounts, filters.Locations, filters.Query)

	// Récupérer tous les artistes
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur API", 500)
		log.Println("❌ Erreur API:", err)
		return
	}

	log.Printf("📊 ÉTAPE 2: %d artistes récupérés depuis l'API", len(allArtists))

	// ÉTAPE 2 : APPLIQUER LE FILTRAGE
	filteredArtists := applyHomeFilters(allArtists, filters)
	log.Printf("🎯 ÉTAPE 2: %d artistes après filtrage (sur %d)", len(filteredArtists), len(allArtists))

	data := HomeData{
		Title:   "Groupie Tracker",
		Artists: filteredArtists, // ← ARTISTES FILTRÉS maintenant !
	}

	templates.Templates.ExecuteTemplate(w, "home.gohtml", data)
}

// parseHomeFilters lit les paramètres URL et les convertit en structure
func parseHomeFilters(r *http.Request) HomeFilters {
	filters := HomeFilters{}

	// Année de création min/max
	if minYear := r.URL.Query().Get("creation_year_min"); minYear != "" {
		if year, err := strconv.Atoi(minYear); err == nil {
			filters.CreationYearMin = year
		}
	}
	if maxYear := r.URL.Query().Get("creation_year_max"); maxYear != "" {
		if year, err := strconv.Atoi(maxYear); err == nil {
			filters.CreationYearMax = year
		}
	}

	// Date premier album min/max
	if minYear := r.URL.Query().Get("album_year_min"); minYear != "" {
		if year, err := strconv.Atoi(minYear); err == nil {
			filters.AlbumYearMin = year
		}
	}
	if maxYear := r.URL.Query().Get("album_year_max"); maxYear != "" {
		if year, err := strconv.Atoi(maxYear); err == nil {
			filters.AlbumYearMax = year
		}
	}

	// Nombre de membres
	memberCountStrs := r.URL.Query()["member_count"]
	for _, countStr := range memberCountStrs {
		if count, err := strconv.Atoi(countStr); err == nil {
			filters.MemberCounts = append(filters.MemberCounts, count)
		}
	}

	// Lieux
	filters.Locations = r.URL.Query()["location"]

	// Query de recherche
	filters.Query = r.URL.Query().Get("q")

	return filters
}

// applyHomeFilters applique les filtres aux artistes
func applyHomeFilters(artists []util.Artist, filters HomeFilters) []util.Artist {
	var results []util.Artist

	for _, artist := range artists {
		// FILTRE 1: Recherche textuelle (nom + membres)
		if filters.Query != "" && !matchesSearchQuery(artist, filters.Query) {
			log.Printf("❌ SEARCH: %s éliminé (ne contient pas '%s')", artist.Name, filters.Query)
			continue
		}

		// FILTRE 2: Année de création
		if filters.CreationYearMin > 0 && artist.CreationDate < filters.CreationYearMin {
			log.Printf("❌ CREATION MIN: %s éliminé (créé en %d < %d)", artist.Name, artist.CreationDate, filters.CreationYearMin)
			continue
		}
		if filters.CreationYearMax > 0 && artist.CreationDate > filters.CreationYearMax {
			log.Printf("❌ CREATION MAX: %s éliminé (créé en %d > %d)", artist.Name, artist.CreationDate, filters.CreationYearMax)
			continue
		}

		// FILTRE 3: Nombre de membres
		if len(filters.MemberCounts) > 0 {
			memberCount := len(artist.Members)
			if !intInHomeSlice(memberCount, filters.MemberCounts) {
				log.Printf("❌ MEMBERS: %s éliminé (%d membres pas dans %v)", artist.Name, memberCount, filters.MemberCounts)
				continue
			}
		}

		// Si on arrive ici, l'artiste passe tous les filtres
		log.Printf("✅ KEEP: %s (créé %d, %d membres)", artist.Name, artist.CreationDate, len(artist.Members))
		results = append(results, artist)
	}

	return results
}

// matchesSearchQuery vérifie si un artiste correspond à la query
func matchesSearchQuery(artist util.Artist, query string) bool {
	queryLower := strings.ToLower(query)

	// Recherche dans le nom
	if strings.Contains(strings.ToLower(artist.Name), queryLower) {
		return true
	}

	// Recherche dans les membres
	for _, member := range artist.Members {
		if strings.Contains(strings.ToLower(member), queryLower) {
			return true
		}
	}

	return false
}

// intInHomeSlice vérifie si un int est dans une slice
func intInHomeSlice(target int, slice []int) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}
