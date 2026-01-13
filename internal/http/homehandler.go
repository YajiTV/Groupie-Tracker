package httphandlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/YajiTV/groupie-tracker/internal/auth"
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

// RelationData représente les relations d'un artiste depuis l'API
type RelationData struct {
	ID        int                 `json:"id"`
	Locations map[string][]string `json:"datesLocations"`
}

// RelationsResponse représente la réponse de l'API relations
type RelationsResponse struct {
	Index []RelationData `json:"index"`
}

// HomeHandler gère la page d'accueil avec filtres
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("HomeHandler appelé avec URL: %s", r.URL.String())

	// Vérifier que c'est bien la route racine
	if r.URL.Path != "/" {
		NotFoundHandler(w, r)
		return
	}

	// Parser les filtres depuis l'URL
	filters := parseHomeFilters(r)
	log.Printf("Filtres détectés: Creation(%d-%d), Albums(%d-%d), Membres%v, Lieux%v, Query:'%s'",
		filters.CreationYearMin, filters.CreationYearMax,
		filters.AlbumYearMin, filters.AlbumYearMax,
		filters.MemberCounts, filters.Locations, filters.Query)

	// Récupérer tous les artistes
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		log.Printf("Erreur API: %v", err)
		return
	}
	log.Printf("%d artistes récupérés depuis l'API", len(allArtists))

	// Récupérer les relations (locations) pour tous les artistes
	artistLocations := fetchArtistLocations()
	log.Printf("%d artistes ont des locations", len(artistLocations))

	// Récupérer tous les lieux disponibles pour le filtre
	allLocations := getAllUniqueLocationsFromRelations(artistLocations)
	log.Printf("%d lieux uniques récupérés", len(allLocations))

	// Appliquer les filtres
	displayedArtists := applyHomeFilters(allArtists, filters, artistLocations)
	log.Printf("%d artistes après filtrage", len(displayedArtists))

	// Préparer les données pour le template
	data := struct {
		Title           string
		Artists         []util.Artist
		Filters         HomeFilters
		AllLocations    []string
		IsAuthenticated bool
	}{
		Title:           "Groupie Tracker",
		Artists:         displayedArtists,
		Filters:         filters,
		AllLocations:    allLocations,
		IsAuthenticated: auth.IsAuthenticated(r),
	}

	// Rendre le template avec gestion d'erreur améliorée
	if err := templates.Templates.ExecuteTemplate(w, "home.gohtml", data); err != nil {
		log.Printf("Erreur template COMPLÈTE: %v", err)
		http.Error(w, fmt.Sprintf("Erreur lors du rendu du template: %v", err), http.StatusInternalServerError)
		return
	}
}

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

	log.Printf("Members parsed: %v from URL params", filters.MemberCounts)
	log.Printf("Locations parsed: %v from URL params", filters.Locations)

	return filters
}

func applyHomeFilters(allArtists []util.Artist, filters HomeFilters, artistLocations map[int][]string) []util.Artist {
	var filteredArtists []util.Artist

	for _, artist := range allArtists {
		// Filtrer par année de création
		if filters.CreationYearMin > 0 && artist.CreationDate < filters.CreationYearMin {
			continue
		}
		if filters.CreationYearMax > 0 && artist.CreationDate > filters.CreationYearMax {
			continue
		}

		// Filtrer par année du premier album
		if filters.AlbumYearMin > 0 && artist.FirstAlbum != "" {
			if len(artist.FirstAlbum) >= 4 {
				firstAlbumYear, err := strconv.Atoi(artist.FirstAlbum[len(artist.FirstAlbum)-4:])
				if err == nil && firstAlbumYear < filters.AlbumYearMin {
					continue
				}
			}
		}
		if filters.AlbumYearMax > 0 && artist.FirstAlbum != "" {
			if len(artist.FirstAlbum) >= 4 {
				firstAlbumYear, err := strconv.Atoi(artist.FirstAlbum[len(artist.FirstAlbum)-4:])
				if err == nil && firstAlbumYear > filters.AlbumYearMax {
					continue
				}
			}
		}

		// Filtrer par nombre de membres
		if len(filters.MemberCounts) > 0 {
			memberMatch := false
			for _, count := range filters.MemberCounts {
				if len(artist.Members) == count {
					memberMatch = true
					break
				}
			}
			if !memberMatch {
				continue
			}
		}

		// NOUVEAU : Filtrer par lieux de concert
		if len(filters.Locations) > 0 {
			locations := artistLocations[artist.ID]
			locationMatch := false

			// Vérifier si l'artiste a joué dans au moins un des lieux sélectionnés
			for _, filterLoc := range filters.Locations {
				for _, artistLoc := range locations {
					if strings.EqualFold(filterLoc, artistLoc) {
						locationMatch = true
						break
					}
				}
				if locationMatch {
					break
				}
			}

			if !locationMatch {
				continue
			}
		}

		// Filtrer par recherche (nom d'artiste, membres, etc.)
		if filters.Query != "" {
			query := strings.ToLower(filters.Query)
			match := false

			// Recherche dans le nom
			if strings.Contains(strings.ToLower(artist.Name), query) {
				match = true
			}

			// Recherche dans les membres
			if !match {
				for _, member := range artist.Members {
					if strings.Contains(strings.ToLower(member), query) {
						match = true
						break
					}
				}
			}

			if !match {
				continue
			}
		}

		filteredArtists = append(filteredArtists, artist)
	}

	return filteredArtists
}

// fetchArtistLocations récupère les locations pour tous les artistes depuis l'API Relations
func fetchArtistLocations() map[int][]string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://groupietrackers.herokuapp.com/api/relation")
	if err != nil {
		log.Printf("Erreur lors de la récupération des relations: %v", err)
		return make(map[int][]string)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Status code non-OK pour relations: %d", resp.StatusCode)
		return make(map[int][]string)
	}

	var relations RelationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&relations); err != nil {
		log.Printf("Erreur lors du décodage des relations: %v", err)
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
	// Ref piscine
	locations := make([]string, 0, len(locationMap))
	for location := range locationMap {
		locations = append(locations, location)
	}

	// AJOUT : Trier par ordre alphabétique
	sort.Strings(locations)

	return locations
}

// Garder pour compatibilité mais maintenant on utilise getAllUniqueLocationsFromRelations
func getAllUniqueLocations() []string {
	artistLocations := fetchArtistLocations()
	return getAllUniqueLocationsFromRelations(artistLocations)
}
