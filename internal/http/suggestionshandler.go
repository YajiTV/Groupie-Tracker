package httphandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/YajiTV/groupie-tracker/internal/util"
)

// Structure for ONE suggestion
type Suggestion struct {
	Text     string `json:"text"`      // The text to display: "Queen" or "Freddie Mercury"
	Type     string `json:"type"`      // The type: "artist" or "member"
	ArtistID int    `json:"artist_id"` // The artist ID to make the link
}

// Structure for the complete response (what we return as JSON)
type SuggestionsResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func SuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get what the user typed
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	// If the search is empty, return an empty list
	if query == "" {
		sendJSONResponse(w, SuggestionsResponse{Suggestions: []Suggestion{}})
		return
	}

	// 2. Retrieve ALL artists from the API
	allArtists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur API", 500)
		log.Println("Suggestions error:", err)
		return
	}

	// 3. Search for matches
	suggestions := findSuggestions(allArtists, query)

	// 4. Return the JSON
	sendJSONResponse(w, SuggestionsResponse{Suggestions: suggestions})
}

// Function that searches for matches in artists
func findSuggestions(artists []util.Artist, query string) []Suggestion {
	var suggestions []Suggestion
	queryLower := strings.ToLower(query)
	maxSuggestions := 8 // Limit to avoid too long list

	for _, artist := range artists {
		if len(suggestions) >= maxSuggestions {
			break
		}

		// SEARCH IN ARTIST NAME
		if strings.Contains(strings.ToLower(artist.Name), queryLower) {
			suggestions = append(suggestions, Suggestion{
				Text:     artist.Name,
				Type:     "artiste",
				ArtistID: artist.ID,
			})
		}

		// SEARCH IN MEMBERS (each member separately)
		for _, member := range artist.Members {
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

// Utility function to return JSON cleanly
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	// Tell the browser we're returning JSON
	w.Header().Set("Content-Type", "application/json")

	// Transform our struct to JSON and send it
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Erreur encodage JSON", 500)
		log.Println("JSON error:", err)
	}
}
