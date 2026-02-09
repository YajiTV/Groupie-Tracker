package util

import (
	"strconv"
	"strings"
)

// Structure for filter parameters
type FilterParams struct {
	CreationYearMin int
	CreationYearMax int
	AlbumYearMin    int
	AlbumYearMax    int
	MemberCounts    []int
	Locations       []string
	Query           string
}

// extractAlbumYear extracts the year from "DD-MM-YYYY" format
func extractAlbumYear(firstAlbum string) int {
	// API format: "14-03-1973"
	parts := strings.Split(firstAlbum, "-")
	if len(parts) != 3 {
		return 0
	}

	// Year is the 3rd element (index 2) after day and month
	if year, err := strconv.Atoi(parts[2]); err == nil {
		return year
	}
	return 0
}

// FilterByCreationYear filters artists by creation year
func FilterByCreationYear(artists []Artist, minYear, maxYear int) []Artist {
	// If no limit, return all artists
	if minYear == 0 && maxYear == 0 {
		return artists
	}

	var filtered []Artist
	for _, artist := range artists {
		// Check limits
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

// ApplyAllFilters applies all filters (for now just creation year)
func ApplyAllFilters(artists []Artist, params FilterParams) []Artist {
	result := artists

	// Filter by creation year
	result = FilterByCreationYear(result, params.CreationYearMin, params.CreationYearMax)

	// TODO: Add other filters gradually

	return result
}
