package httphandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/YajiTV/groupie-tracker/internal/util"
)

// Structure pour UNE suggestion
type Suggestion struct {
	Text     string `json:"text"`      // Le texte à afficher : "Queen" ou "Freddie Mercury"
	Type     string `json:"type"`      // Le type : "artiste" ou "membre"
	ArtistID int    `json:"artist_id"` // L'ID de l'artiste pour faire le lien
}

// Structure pour la réponse complète (ce qu'on renvoie en JSON)
type SuggestionsResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func SuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Récupérer ce que l'utilisateur a tapé
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	// Si la recherche est vide, on renvoie une liste vide
	if query == "" {
		sendJSONResponse(w, SuggestionsResponse{Suggestions: []Suggestion{}})
		return
	}

	// 2. Récupérer TOUS les artistes de l'API
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur API", 500)
		log.Println("Erreur suggestions:", err)
		return
	}

	// 3. Chercher les correspondances
	suggestions := findSuggestions(allArtists, query)

	// 4. Renvoyer le JSON
	sendJSONResponse(w, SuggestionsResponse{Suggestions: suggestions})
}

// Fonction qui cherche les correspondances dans les artistes
func findSuggestions(artists []util.Artist, query string) []Suggestion {
	var suggestions []Suggestion
	queryLower := strings.ToLower(query)
	maxSuggestions := 8 // Limite à 8 suggestions max

	for _, artist := range artists {
		// Si on a déjà assez de suggestions, on s'arrête
		if len(suggestions) >= maxSuggestions {
			break
		}

		// RECHERCHE DANS LE NOM DE L'ARTISTE
		if strings.Contains(strings.ToLower(artist.Name), queryLower) {
			suggestions = append(suggestions, Suggestion{
				Text:     artist.Name,
				Type:     "artiste",
				ArtistID: artist.ID,
			})
		}

		// RECHERCHE DANS LES MEMBRES (chaque membre séparément)
		for _, member := range artist.Members {
			// Vérifier qu'on n'a pas déjà trop de suggestions
			if len(suggestions) >= maxSuggestions {
				break
			}

			if strings.Contains(strings.ToLower(member), queryLower) {
				suggestions = append(suggestions, Suggestion{
					Text:     member,
					Type:     "membre",
					ArtistID: artist.ID,
				})
			}
		}
	}

	return suggestions
}

// Fonction utilitaire pour renvoyer du JSON proprement
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	// Dire au navigateur qu'on renvoie du JSON
	w.Header().Set("Content-Type", "application/json")

	// Transformer notre struct en JSON et l'envoyer
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Erreur encodage JSON", 500)
		log.Println("Erreur JSON:", err)
	}
}
