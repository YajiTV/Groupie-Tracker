package httphandlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/YajiTV/groupie-tracker/internal/util"
)

// HomeFilters represents the home page filters
type HomeFilters struct {
	CreationYearMin int
	CreationYearMax int
	AlbumYearMin    int
	AlbumYearMax    int
	MemberCounts    []int
	Locations       []string
	Query           string
}

// RelationData represents an artist's relations from the API
type RelationData struct {
	ID        int                 `json:"id"`
	Locations map[string][]string `json:"datesLocations"`
}

// RelationsResponse represents the relations API response
type RelationsResponse struct {
	Index []RelationData `json:"index"`
}

// parseHomeFilters extracts and parses filters from URL parameters
func parseHomeFilters(r *http.Request) HomeFilters {
	query := r.URL.Query()
	filters := HomeFilters{}

	// Parse creation years
	if val := query.Get("creation_year_min"); val != "" {
		filters.CreationYearMin, _ = strconv.Atoi(val)
	}
	if val := query.Get("creation_year_max"); val != "" {
		filters.CreationYearMax, _ = strconv.Atoi(val)
	}

	// Parse album years
	if val := query.Get("album_year_min"); val != "" {
		filters.AlbumYearMin, _ = strconv.Atoi(val)
	}
	if val := query.Get("album_year_max"); val != "" {
		filters.AlbumYearMax, _ = strconv.Atoi(val)
	}

	// Parse member counts
	if memberCounts := query["member_count"]; len(memberCounts) > 0 {
		memberMap := make(map[int]bool) // Map for automatic deduplication
		for _, m := range memberCounts {
			if count, err := strconv.Atoi(strings.TrimSpace(m)); err == nil && count > 0 {
				memberMap[count] = true
			}
		}
		for count := range memberMap {
			filters.MemberCounts = append(filters.MemberCounts, count)
		}
	}

	// Parse locations
	if locations := query["location"]; len(locations) > 0 {
		for _, loc := range locations {
			cleanLoc := strings.TrimSpace(loc)
			if cleanLoc != "" {
				filters.Locations = append(filters.Locations, cleanLoc)
			}
		}
	}

	// Parse search query
	filters.Query = strings.TrimSpace(query.Get("q"))

	return filters
}

// applyHomeFilters applies filters on the artist list
func applyHomeFilters(allArtists []util.Artist, filters HomeFilters, artistLocations map[int][]string) []util.Artist {
	var filteredArtists []util.Artist

	for _, artist := range allArtists {
		// Filter by creation year
		if !filterByCreationYear(artist, filters) {
			continue
		}

		// Filter by first album year
		if !filterByFirstAlbum(artist, filters) {
			continue
		}

		// Filter by member count
		if !filterByMemberCount(artist, filters) {
			continue
		}

		// Filter by concert locations
		if !filterByLocation(artist, filters, artistLocations) {
			continue
		}

		// Filter by text search
		if !filterByQuery(artist, filters) {
			continue
		}

		filteredArtists = append(filteredArtists, artist)
	}

	return filteredArtists
}

// filterByCreationYear checks if the artist matches the creation year filter
func filterByCreationYear(artist util.Artist, filters HomeFilters) bool {
	if filters.CreationYearMin > 0 && artist.CreationDate < filters.CreationYearMin {
		return false
	}
	if filters.CreationYearMax > 0 && artist.CreationDate > filters.CreationYearMax {
		return false
	}
	return true
}

// filterByFirstAlbum checks if the artist matches the first album year filter
func filterByFirstAlbum(artist util.Artist, filters HomeFilters) bool {
	if artist.FirstAlbum == "" || len(artist.FirstAlbum) < 4 {
		return true
	}

	firstAlbumYear, err := strconv.Atoi(artist.FirstAlbum[len(artist.FirstAlbum)-4:]) // Extract last 4 characters (year)
	if err != nil {
		return true
	}

	if filters.AlbumYearMin > 0 && firstAlbumYear < filters.AlbumYearMin {
		return false
	}
	if filters.AlbumYearMax > 0 && firstAlbumYear > filters.AlbumYearMax {
		return false
	}
	return true
}

// filterByMemberCount checks if the artist matches the member count filter
func filterByMemberCount(artist util.Artist, filters HomeFilters) bool {
	if len(filters.MemberCounts) == 0 {
		return true
	}

	for _, count := range filters.MemberCounts {
		if len(artist.Members) == count {
			return true
		}
	}
	return false
}

// filterByLocation checks if the artist matches the concert location filter
func filterByLocation(artist util.Artist, filters HomeFilters, artistLocations map[int][]string) bool {
	if len(filters.Locations) == 0 {
		return true
	}

	locations := artistLocations[artist.ID]

	for _, filterLoc := range filters.Locations {
		for _, artistLoc := range locations {
			if strings.EqualFold(filterLoc, artistLoc) {
				return true
			}
		}
	}

	return false
}

// filterByQuery checks if the artist matches the text search
func filterByQuery(artist util.Artist, filters HomeFilters) bool {
	if filters.Query == "" {
		return true
	}

	query := strings.ToLower(filters.Query)

	// Search in name
	if strings.Contains(strings.ToLower(artist.Name), query) {
		return true
	}

	// Search in members
	for _, member := range artist.Members {
		if strings.Contains(strings.ToLower(member), query) {
			return true
		}
	}

	return false
}

// fetchArtistLocations retrieves locations for all artists from the Relations API
func fetchArtistLocations() map[int][]string {
	client := &http.Client{
		Timeout: 10 * time.Second, // Timeout to avoid blocking
	}

	resp, err := client.Get("https://groupietrackers.herokuapp.com/api/relation")
	if err != nil {
		return make(map[int][]string)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return make(map[int][]string)
	}

	var relations RelationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&relations); err != nil {
		return make(map[int][]string)
	}

	artistLocations := make(map[int][]string)

	for _, relation := range relations.Index {
		var locations []string
		for location := range relation.Locations {
			cleanLocation := strings.TrimSpace(location)
			if cleanLocation != "" {
				locations = append(locations, cleanLocation)
			}
		}
		artistLocations[relation.ID] = locations
	}

	return artistLocations
}

// getAllUniqueLocationsFromRelations extracts all unique locations from relations
func getAllUniqueLocationsFromRelations(artistLocations map[int][]string) []string {
	locationMap := make(map[string]bool) // Map for automatic deduplication

	for _, locations := range artistLocations {
		for _, location := range locations {
			locationMap[location] = true
		}
	}

	locations := make([]string, 0, len(locationMap))
	for location := range locationMap {
		locations = append(locations, location)
	}

	sort.Slice(locations, func(i, j int) bool {
		return strings.ToLower(locations[i]) < strings.ToLower(locations[j])
	})

	return locations
}

// getAllUniqueLocations retrieves all unique locations (wrapper function for compatibility)
func getAllUniqueLocations() []string {
	artistLocations := fetchArtistLocations()
	return getAllUniqueLocationsFromRelations(artistLocations)
}
