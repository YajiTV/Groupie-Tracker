package httphandlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

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
		log.Printf("❌ Erreur API: %v", err)
		return
	}
	log.Printf("📊 %d artistes récupérés depuis l'API", len(allArtists))

	// Appliquer les filtres
	displayedArtists := applyHomeFilters(allArtists, filters)
	log.Printf("✅ %d artistes envoyés au template", len(displayedArtists))

	// Préparer les données pour le template
	data := struct {
		Title   string
		Artists []util.Artist
	}{
		Title:   "Groupie Tracker",
		Artists: displayedArtists,
	}

	// Rendre le template
	if err := templates.Templates.ExecuteTemplate(w, "home.gohtml", data); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
		log.Printf("❌ Erreur template: %v", err)
	}
}

// parseHomeFilters extrait les filtres des paramètres URL
func parseHomeFilters(r *http.Request) HomeFilters {
	filters := HomeFilters{}

	// Année de création min
	if yearStr := r.URL.Query().Get("creation_year_min"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filters.CreationYearMin = year
		}
	}

	// Année de création max
	if yearStr := r.URL.Query().Get("creation_year_max"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filters.CreationYearMax = year
		}
	}

	// Année premier album min
	if yearStr := r.URL.Query().Get("album_year_min"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filters.AlbumYearMin = year
		}
	}

	// Année premier album max
	if yearStr := r.URL.Query().Get("album_year_max"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filters.AlbumYearMax = year
		}
	}

	// Nombre de membres (peut avoir plusieurs valeurs)
	if memberStrs := r.URL.Query()["member_count"]; len(memberStrs) > 0 {
		for _, memberStr := range memberStrs {
			if member, err := strconv.Atoi(memberStr); err == nil {
				filters.MemberCounts = append(filters.MemberCounts, member)
			}
		}
	}

	// Lieux (peut avoir plusieurs valeurs)
	filters.Locations = r.URL.Query()["location"]

	// Recherche textuelle
	filters.Query = r.URL.Query().Get("q")

	return filters
}

// applyHomeFilters applique les filtres sur la liste des artistes
func applyHomeFilters(artists []util.Artist, filters HomeFilters) []util.Artist {
	var filtered []util.Artist

	for _, artist := range artists {
		// Filtre par recherche textuelle (nom + membres)
		if filters.Query != "" {
			if !matchesSearchQuery(artist, filters.Query) {
				log.Printf("❌ SEARCH: %s éliminé (ne contient pas '%s')", artist.Name, filters.Query)
				continue
			}
		}

		// Filtre par année de création min
		if filters.CreationYearMin > 0 {
			if artist.CreationDate < filters.CreationYearMin {
				log.Printf("❌ CREATION MIN: %s éliminé (année %d < %d)", artist.Name, artist.CreationDate, filters.CreationYearMin)
				continue
			}
		}

		// Filtre par année de création max
		if filters.CreationYearMax > 0 {
			if artist.CreationDate > filters.CreationYearMax {
				log.Printf("❌ CREATION MAX: %s éliminé (année %d > %d)", artist.Name, artist.CreationDate, filters.CreationYearMax)
				continue
			}
		}

		// Filtre par nombre de membres
		if len(filters.MemberCounts) > 0 {
			memberCount := len(artist.Members)
			if !intInHomeSlice(memberCount, filters.MemberCounts) {
				log.Printf("❌ MEMBERS: %s éliminé (%d membres pas dans %v)", artist.Name, memberCount, filters.MemberCounts)
				continue
			}
		}

		// Si l'artiste passe tous les filtres, l'ajouter
		log.Printf("✅ KEEP: %s (année: %d, membres: %d)", artist.Name, artist.CreationDate, len(artist.Members))
		filtered = append(filtered, artist)
	}

	log.Printf("🎯 Résultat final: %d artistes après filtrage", len(filtered))
	return filtered
}

// matchesSearchQuery vérifie si un artiste correspond à la recherche
func matchesSearchQuery(artist util.Artist, query string) bool {
	queryLower := strings.ToLower(query)

	// Recherche dans le nom de l'artiste
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

// intInHomeSlice vérifie si un entier est présent dans un slice
func intInHomeSlice(target int, slice []int) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}
