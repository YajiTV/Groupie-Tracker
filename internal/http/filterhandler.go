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

// RelationData représente les relations d'un artiste depuis l'API
type RelationData struct {
	ID        int                 `json:"id"`
	Locations map[string][]string `json:"datesLocations"`
}

// RelationsResponse représente la réponse de l'API relations
type RelationsResponse struct {
	Index []RelationData `json:"index"`
}

// parseHomeFilters extrait et parse les filtres depuis les paramètres URL
func parseHomeFilters(r *http.Request) HomeFilters {
	query := r.URL.Query()
	filters := HomeFilters{}

	// Parser les années de création
	if val := query.Get("creation_year_min"); val != "" {
		filters.CreationYearMin, _ = strconv.Atoi(val)
	}
	if val := query.Get("creation_year_max"); val != "" {
		filters.CreationYearMax, _ = strconv.Atoi(val)
	}

	// Parser les années d'album
	if val := query.Get("album_year_min"); val != "" {
		filters.AlbumYearMin, _ = strconv.Atoi(val)
	}
	if val := query.Get("album_year_max"); val != "" {
		filters.AlbumYearMax, _ = strconv.Atoi(val)
	}

	// Parser les nombres de membres
	if memberCounts := query["member_count"]; len(memberCounts) > 0 {
		memberMap := make(map[int]bool)
		for _, m := range memberCounts {
			if count, err := strconv.Atoi(strings.TrimSpace(m)); err == nil && count > 0 {
				memberMap[count] = true
			}
		}
		for count := range memberMap {
			filters.MemberCounts = append(filters.MemberCounts, count)
		}
	}

	// Parser les lieux
	if locations := query["location"]; len(locations) > 0 {
		for _, loc := range locations {
			cleanLoc := strings.TrimSpace(loc)
			if cleanLoc != "" {
				filters.Locations = append(filters.Locations, cleanLoc)
			}
		}
	}

	// Parser la recherche
	filters.Query = strings.TrimSpace(query.Get("q"))

	return filters
}

// applyHomeFilters applique les filtres sur la liste d'artistes
func applyHomeFilters(allArtists []util.Artist, filters HomeFilters, artistLocations map[int][]string) []util.Artist {
	var filteredArtists []util.Artist

	for _, artist := range allArtists {
		// Filtrer par année de création
		if !filterByCreationYear(artist, filters) {
			continue
		}

		// Filtrer par année du premier album
		if !filterByFirstAlbum(artist, filters) {
			continue
		}

		// Filtrer par nombre de membres
		if !filterByMemberCount(artist, filters) {
			continue
		}

		// Filtrer par lieux de concert
		if !filterByLocation(artist, filters, artistLocations) {
			continue
		}

		// Filtrer par recherche textuelle
		if !filterByQuery(artist, filters) {
			continue
		}

		filteredArtists = append(filteredArtists, artist)
	}

	return filteredArtists
}

// filterByCreationYear vérifie si l'artiste correspond au filtre d'année de création
func filterByCreationYear(artist util.Artist, filters HomeFilters) bool {
	if filters.CreationYearMin > 0 && artist.CreationDate < filters.CreationYearMin {
		return false
	}
	if filters.CreationYearMax > 0 && artist.CreationDate > filters.CreationYearMax {
		return false
	}
	return true
}

// filterByFirstAlbum vérifie si l'artiste correspond au filtre d'année du premier album
func filterByFirstAlbum(artist util.Artist, filters HomeFilters) bool {
	if artist.FirstAlbum == "" || len(artist.FirstAlbum) < 4 {
		return true
	}

	firstAlbumYear, err := strconv.Atoi(artist.FirstAlbum[len(artist.FirstAlbum)-4:])
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

// filterByMemberCount vérifie si l'artiste correspond au filtre de nombre de membres
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

// filterByLocation vérifie si l'artiste correspond au filtre de lieux de concert
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

// filterByQuery vérifie si l'artiste correspond à la recherche textuelle
func filterByQuery(artist util.Artist, filters HomeFilters) bool {
	if filters.Query == "" {
		return true
	}

	query := strings.ToLower(filters.Query)

	// Recherche dans le nom
	if strings.Contains(strings.ToLower(artist.Name), query) {
		return true
	}

	// Recherche dans les membres
	for _, member := range artist.Members {
		if strings.Contains(strings.ToLower(member), query) {
			return true
		}
	}

	return false
}

// fetchArtistLocations récupère les locations pour tous les artistes depuis l'API Relations
func fetchArtistLocations() map[int][]string {
	client := &http.Client{
		Timeout: 10 * time.Second,
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

	// Construire la map artistID -> []locations
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

// getAllUniqueLocationsFromRelations extrait tous les lieux uniques depuis les relations
func getAllUniqueLocationsFromRelations(artistLocations map[int][]string) []string {
	locationMap := make(map[string]bool)

	for _, locations := range artistLocations {
		for _, location := range locations {
			locationMap[location] = true
		}
	}

	// Convertir la map en slice
	locations := make([]string, 0, len(locationMap))
	for location := range locationMap {
		locations = append(locations, location)
	}

	// Trier par ordre alphabétique (insensible à la casse)
	sort.Slice(locations, func(i, j int) bool {
		return strings.ToLower(locations[i]) < strings.ToLower(locations[j])
	})

	return locations
}

// getAllUniqueLocations récupère tous les lieux uniques (fonction wrapper pour compatibilité)
func getAllUniqueLocations() []string {
	artistLocations := fetchArtistLocations()
	return getAllUniqueLocationsFromRelations(artistLocations)
}
