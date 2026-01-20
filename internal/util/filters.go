package util

import (
	"strconv"
	"strings"
)

// Structure pour les paramètres de filtres
type FilterParams struct {
	CreationYearMin int
	CreationYearMax int
	AlbumYearMin    int
	AlbumYearMax    int
	MemberCounts    []int
	Locations       []string
	Query           string
}

// extractAlbumYear extrait l'année du format "DD-MM-YYYY"
func extractAlbumYear(firstAlbum string) int {
	// Format API: "14-03-1973"
	parts := strings.Split(firstAlbum, "-")
	if len(parts) != 3 {
		return 0 // Données invalides
	}

	// L'année est le 3ème élément (index 2)
	if year, err := strconv.Atoi(parts[2]); err == nil {
		return year
	}
	return 0
}

// FilterByCreationYear filtre les artistes par année de création
func FilterByCreationYear(artists []Artist, minYear, maxYear int) []Artist {
	// Si aucune limite, retourner tous les artistes
	if minYear == 0 && maxYear == 0 {
		return artists
	}

	var filtered []Artist
	for _, artist := range artists {
		// Vérifier les limites
		if minYear > 0 && artist.CreationDate < minYear {
			continue
		}
		if maxYear > 0 && artist.CreationDate > maxYear {
			continue
		}

		filtered = append(filtered, artist)
	}
	return filtered
}

// ApplyAllFilters applique tous les filtres (pour l'instant juste année de création)
func ApplyAllFilters(artists []Artist, params FilterParams) []Artist {
	result := artists

	// Filtrage par année de création
	result = FilterByCreationYear(result, params.CreationYearMin, params.CreationYearMax)

	// TODO: Ajouter les autres filtres au fur et à mesure

	return result
}
