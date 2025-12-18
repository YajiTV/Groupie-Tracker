package httphandlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

type FiltersData struct {
	Title            string
	Artists          []util.Artist
	Count            int
	FilterOptions    FilterOptions
	AppliedFilters   AppliedFilters
	HasActiveFilters bool
}

type FilterOptions struct {
	MinCreationYear int
	MaxCreationYear int
	MinAlbumYear    int
	MaxAlbumYear    int
	MemberCounts    []int
	Locations       []string
}

type AppliedFilters struct {
	CreationYearMin int
	CreationYearMax int
	AlbumYearMin    int
	AlbumYearMax    int
	MemberCounts    []int
	Locations       []string
	Query           string
}

func FiltersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🎯 ÉTAPE 1: Début - affichage de tous les artistes")

	// Récupérer tous les artistes
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur API", 500)
		log.Println("❌ Erreur lors de la récupération des artistes:", err)
		return
	}

	log.Printf("📊 ÉTAPE 1: %d artistes récupérés depuis l'API", len(allArtists))

	// Pour l'instant, on affiche tous les artistes SANS AUCUN FILTRE
	data := FiltersData{
		Title:   "Tous les artistes (sans filtres)",
		Artists: allArtists,
		Count:   len(allArtists),
	}

	templates.Templates.ExecuteTemplate(w, "filters.gohtml", data)
	log.Printf("✅ ÉTAPE 1: Page envoyée avec %d artistes", len(allArtists))
}

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

		// Filtre nombre de membres
		if len(filters.MemberCounts) > 0 {
			memberCount := len(artist.Members)
			if !intInSlice(memberCount, filters.MemberCounts) {
				continue
			}
		}

		// Filtre lieux (avec les données de relations)
		if len(filters.Locations) > 0 {
			if !artistHasLocation(artist, filters.Locations, relationResponse) {
				continue
			}
		}

		results = append(results, artist)
	}

	return results
}

func generateFilterOptions(artists []util.Artist, locationResponse util.LocationResponse) FilterOptions {
	options := FilterOptions{
		MinCreationYear: 9999,
		MaxCreationYear: 0,
		MinAlbumYear:    9999,
		MaxAlbumYear:    0,
	}

	memberCountSet := make(map[int]bool)
	locationSet := make(map[string]bool)

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
	// Trier les nombres de membres
	for i := 0; i < len(options.MemberCounts); i++ {
		for j := i + 1; j < len(options.MemberCounts); j++ {
			if options.MemberCounts[i] > options.MemberCounts[j] {
				options.MemberCounts[i], options.MemberCounts[j] = options.MemberCounts[j], options.MemberCounts[i]
			}
		}
	}

	// Extraire les lieux uniques
	for _, locData := range locationResponse.Index {
		for _, location := range locData.Locations {
			locationSet[location] = true
		}
	}

	for location := range locationSet {
		options.Locations = append(options.Locations, location)
	}

	// Trier les locations alphabétiquement
	for i := 0; i < len(options.Locations); i++ {
		for j := i + 1; j < len(options.Locations); j++ {
			if options.Locations[i] > options.Locations[j] {
				options.Locations[i], options.Locations[j] = options.Locations[j], options.Locations[i]
			}
		}
	}

	log.Printf("⚙️  FILTERS: Options générées - Years: %d-%d, Albums: %d-%d, Members: %v, Locations: %d",
		options.MinCreationYear, options.MaxCreationYear,
		options.MinAlbumYear, options.MaxAlbumYear,
		options.MemberCounts, len(options.Locations))

	return options
}

// Fonctions utilitaires
func extractAlbumYear(firstAlbum string) int {
	// Format: "DD-MM-YYYY"
	parts := strings.Split(firstAlbum, "-")
	if len(parts) != 3 {
		return 0
	}
	if year, err := strconv.Atoi(parts[2]); err == nil {
		return year
	}
	return 0
}

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

func intInSlice(target int, slice []int) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

func artistHasLocation(artist util.Artist, locations []string, relationResponse util.RelationResponse) bool {
	// Si aucune location sélectionnée, retourner true
	if len(locations) == 0 {
		return true
	}

	// Trouver les relations pour cet artiste
	for _, relData := range relationResponse.Index {
		if relData.ID == artist.ID {
			// Vérifier si l'artiste a joué dans l'un des lieux sélectionnés
			for _, selectedLocation := range locations {
				// Parcourir toutes les locations dans datesLocations
				for locationKey := range relData.DatesLocations {
					// Les clés sont formatées comme "city-country", on compare avec selectedLocation
					if strings.Contains(strings.ToLower(locationKey), strings.ToLower(selectedLocation)) ||
						strings.Contains(strings.ToLower(selectedLocation), strings.ToLower(locationKey)) {
						log.Printf("✅ LOCATION: %s trouvé dans %s", artist.Name, locationKey)
						return true
					}
				}
			}
			log.Printf("❌ LOCATION: %s non trouvé dans %v", artist.Name, locations)
			break
		}
	}
	return false
}

func hasActiveFilters(filters AppliedFilters) bool {
	return filters.CreationYearMin > 0 || filters.CreationYearMax > 0 ||
		filters.AlbumYearMin > 0 || filters.AlbumYearMax > 0 ||
		len(filters.MemberCounts) > 0 || len(filters.Locations) > 0 ||
		filters.Query != ""
}
