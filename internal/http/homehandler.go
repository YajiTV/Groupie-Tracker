package httphandlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// Récupérer tous les lieux disponibles pour le filtre
	allLocations := getAllUniqueLocations()
	log.Printf("%d lieux uniques récupérés", len(allLocations))

	// Appliquer les filtres
	displayedArtists := applyHomeFilters(allArtists, filters)
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
		http.Error(w, fmt.Sprintf("Erreur lors du rendu du template: %v", err), http.StatusInternalServerError)
		return
	}
}

func parseHomeFilters(r *http.Request) HomeFilters {
	query := r.URL.Query()
	filters := HomeFilters{}

	// CORRECTION : Parser avec les bons noms de paramètres
	// L'URL contient creation_year_min, album_year_min, etc.
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

	// Parser les nombres de membres (plusieurs paramètres possibles)
	// Soit un range avec member_count_min et member_count_max
	memberCountMin := 0
	memberCountMax := 0

	if val := query.Get("member_count_min"); val != "" {
		memberCountMin, _ = strconv.Atoi(val)
	}
	if val := query.Get("member_count_max"); val != "" {
		memberCountMax, _ = strconv.Atoi(val)
	}

	// Générer la liste des nombres de membres dans le range
	if memberCountMin > 0 && memberCountMax > 0 {
		for i := memberCountMin; i <= memberCountMax; i++ {
			filters.MemberCounts = append(filters.MemberCounts, i)
		}
	}

	// Ou des valeurs spécifiques séparées par des virgules
	if val := query.Get("members"); val != "" {
		members := strings.Split(val, ",")
		filters.MemberCounts = []int{} // Reset si on a des valeurs spécifiques
		for _, m := range members {
			if count, err := strconv.Atoi(strings.TrimSpace(m)); err == nil {
				filters.MemberCounts = append(filters.MemberCounts, count)
			}
		}
	}

	// Parser les lieux
	if val := query.Get("locations"); val != "" {
		filters.Locations = strings.Split(val, ",")
		for i := range filters.Locations {
			filters.Locations[i] = strings.TrimSpace(filters.Locations[i])
		}
	}

	// Parser la recherche
	filters.Query = strings.TrimSpace(query.Get("q"))

	return filters
}

func applyHomeFilters(allArtists []util.Artist, filters HomeFilters) []util.Artist {
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

func getAllUniqueLocations() []string {
	// Structure pour l'API Relations de Groupie Tracker
	type RelationData struct {
		Index     int                 `json:"index"`
		Locations map[string][]string `json:"datesLocations"`
	}

	type RelationsResponse struct {
		Index []RelationData `json:"index"`
	}

	// Récupérer les relations depuis l'API avec timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://groupietrackers.herokuapp.com/api/relation")
	if err != nil {
		log.Printf("Erreur lors de la récupération des relations: %v", err)
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Status code non-OK pour relations: %d", resp.StatusCode)
		return []string{}
	}

	var relations RelationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&relations); err != nil {
		log.Printf("Erreur lors du décodage des relations: %v", err)
		return []string{}
	}

	// Extraire tous les lieux uniques
	locationMap := make(map[string]bool)

	for _, relation := range relations.Index {
		for location := range relation.Locations {
			// Nettoyer et normaliser le nom du lieu
			cleanLocation := strings.TrimSpace(location)
			if cleanLocation != "" {
				locationMap[cleanLocation] = true
			}
		}
	}

	// Convertir la map en slice
	locations := make([]string, 0, len(locationMap))
	for location := range locationMap {
		locations = append(locations, location)
	}

	return locations
}
