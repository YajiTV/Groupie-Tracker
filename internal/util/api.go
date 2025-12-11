package util

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const ArtistsURL = "https://groupietrackers.herokuapp.com/api/artists"

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
}

func (a Artist) GetSpotifyURL() string {
	// Encode le nom de l'artiste pour l'URL
	encodedName := url.QueryEscape(a.Name)
	return "https://open.spotify.com/search/" + encodedName
}

func FetchArtists() ([]Artist, error) {
	resp, err := http.Get(ArtistsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artists []Artist
	json.NewDecoder(resp.Body).Decode(&artists)
	return artists, nil
}

func FetchArtistByID(id int) (Artist, error) {
	resp, err := http.Get(ArtistsURL)
	if err != nil {
		return Artist{}, err
	}
	defer resp.Body.Close()

	var artists []Artist
	json.NewDecoder(resp.Body).Decode(&artists)

	for _, artist := range artists {
		if artist.ID == id {
			return artist, nil
		}
	}
	return Artist{}, err
}
