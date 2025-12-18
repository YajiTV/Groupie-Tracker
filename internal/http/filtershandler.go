package httphandlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

// FiltersData contient les données pour la page de filtres
type FiltersData struct {
	Title            string
	Artists          []util.Artist
	Count            int
	FilterOptions    FilterOptions
	AppliedFilters   AppliedFilters
	HasActiveFilters bool
}

// FilterOptions contient les options disponibles pour les filtres
type FilterOptions struct {
	MinCreationYear int
	MaxCreationYear int
	MinAlbumYear    int
	MaxAlbumYear    int
	MemberCounts    []int
	Locations       []string
}

// AppliedFilters contient les filtres appliqués par l'utilisateur
type AppliedFilters struct {
	CreationYearMin int
	CreationYearMax int
	AlbumYearMin    int
	AlbumYearMax    int
	MemberCountMin  int
	MemberCountMax  int
	Query           string
}

// FiltersHandler gère la page de filtres dédiée
func FiltersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🎯 Chargement de la page des filtres")

	// Récupérer tous les artistes
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur API", http.StatusInternalServerError)
		log.Println("❌ Erreur lors de la récupération des artistes:", err)
		return
	}

	// Récupérer les locations et relations
	locationResponse, err := util.FetchLocations()
	if err != nil {
		log.Println("⚠️  Erreur lors de la récupération des locations:", err)
	}

	relationResponse, err := util.FetchRelations()
	if err != nil {
		log.Println("⚠️  Erreur lors de la récupération des relations:", err)
	}

	// Parser les filtres depuis l'URL
	appliedFilters := parseFilters(r)

	// Appliquer les filtres
	filteredArtists := applyFilters(allArtists, appliedFilters, relationResponse)

	// Générer les options de filtres
	filterOptions := generateFilterOptions(allArtists, locationResponse)

	// Préparer les données pour le template
	data := FiltersData{
		Title:            "Filtres avancés",
		Artists:          filteredArtists,
		Count:            len(filteredArtists),
		FilterOptions:    filterOptions,
		AppliedFilters:   appliedFilters,
		HasActiveFilters: hasActiveFilters(appliedFilters),
	}

	if err := templates.Templates.ExecuteTemplate(w, "filters.gohtml", data); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
		log.Printf("❌ Erreur template: %v", err)
	}

	log.Printf("✅ Page filtres envoyée : %d artistes affichés", len(filteredArtists))
}

// parseFilters extrait les paramètres de filtres depuis la requête
func parseFilters(r *http.Request) AppliedFilters {
	filters := AppliedFilters{}

	// Année de création
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

	// Date premier album
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

	// Nombre de membres (curseur min/max)
	if minCount := r.URL.Query().Get("member_count_min"); minCount != "" {
		if count, err := strconv.Atoi(minCount); err == nil {
			filters.MemberCountMin = count
		}
	}
	if maxCount := r.URL.Query().Get("member_count_max"); maxCount != "" {
		if count, err := strconv.Atoi(maxCount); err == nil {
			filters.MemberCountMax = count
		}
	}

	// Query de recherche
	filters.Query = r.URL.Query().Get("q")

	return filters
}

// applyFilters applique tous les filtres sur la liste des artistes
func applyFilters(artists []util.Artist, filters AppliedFilters, relationResponse util.RelationResponse) []util.Artist {
	var results []util.Artist

	for _, artist := range artists {
		// Recherche textuelle (si présente)
		if filters.Query != "" && !matchesQuery(artist, filters.Query) {
			continue
		}

		// Filtre année de création
		if filters.CreationYearMin > 0 && artist.CreationDate < filters.CreationYearMin {
			continue
		}
		if filters.CreationYearMax > 0 && artist.CreationDate > filters.CreationYearMax {
			continue
		}

		// Filtre date premier album
		albumYear := extractAlbumYear(artist.FirstAlbum)
		if filters.AlbumYearMin > 0 && albumYear < filters.AlbumYearMin {
			continue
		}
		if filters.AlbumYearMax > 0 && albumYear > filters.AlbumYearMax {
			continue
		}

		// Filtre nombre de membres (curseur min/max)
		memberCount := len(artist.Members)
		if filters.MemberCountMin > 0 && memberCount < filters.MemberCountMin {
			continue
		}
		if filters.MemberCountMax > 0 && memberCount > filters.MemberCountMax {
			continue
		}

		results = append(results, artist)
	}

	return results
}

// generateFilterOptions génère les options disponibles pour les filtres
func generateFilterOptions(artists []util.Artist, locationResponse util.LocationResponse) FilterOptions {
	options := FilterOptions{
		MinCreationYear: 9999,
		MaxCreationYear: 0,
		MinAlbumYear:    9999,
		MaxAlbumYear:    0,
	}

	memberCountSet := make(map[int]bool)

	// Analyser tous les artistes pour les options
	for _, artist := range artists {
		// Années de création
		if artist.CreationDate < options.MinCreationYear {
			options.MinCreationYear = artist.CreationDate
		}
		if artist.CreationDate > options.MaxCreationYear {
			options.MaxCreationYear = artist.CreationDate
		}

		// Années premier album
		albumYear := extractAlbumYear(artist.FirstAlbum)
		if albumYear > 0 {
			if albumYear < options.MinAlbumYear {
				options.MinAlbumYear = albumYear
			}
			if albumYear > options.MaxAlbumYear {
				options.MaxAlbumYear = albumYear
			}
		}

		// Nombre de membres
		memberCountSet[len(artist.Members)] = true
	}

	// Convertir les sets en slices et trier
	for count := range memberCountSet {
		options.MemberCounts = append(options.MemberCounts, count)
	}

	// Trier les nombres de membres (bubble sort simple)
	for i := 0; i < len(options.MemberCounts); i++ {
		for j := i + 1; j < len(options.MemberCounts); j++ {
			if options.MemberCounts[i] > options.MemberCounts[j] {
				options.MemberCounts[i], options.MemberCounts[j] = options.MemberCounts[j], options.MemberCounts[i]
			}
		}
	}

	log.Printf("⚙️  Options générées - Création: %d-%d, Albums: %d-%d, Membres: %v",
		options.MinCreationYear, options.MaxCreationYear,
		options.MinAlbumYear, options.MaxAlbumYear,
		options.MemberCounts)

	return options
}

// extractAlbumYear extrait l'année du format "DD-MM-YYYY"
func extractAlbumYear(firstAlbum string) int {
	parts := strings.Split(firstAlbum, "-")
	if len(parts) != 3 {
		return 0
	}
	if year, err := strconv.Atoi(parts[2]); err == nil {
		return year
	}
	return 0
}

// matchesQuery vérifie si un artiste correspond à la recherche textuelle
func matchesQuery(artist util.Artist, query string) bool {
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

// hasActiveFilters vérifie si des filtres sont actifs
func hasActiveFilters(filters AppliedFilters) bool {
	return filters.CreationYearMin > 0 || filters.CreationYearMax > 0 ||
		filters.AlbumYearMin > 0 || filters.AlbumYearMax > 0 ||
		filters.MemberCountMin > 0 || filters.MemberCountMax > 0 ||
		filters.Query != ""
}
