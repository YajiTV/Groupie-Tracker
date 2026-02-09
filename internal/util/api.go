package util

import (
	"encoding/json"
	"net/http"
	"net/url"
	//"strconv"
)

const (
	ArtistsURL   = "https://groupietrackers.herokuapp.com/api/artists"
	LocationsURL = "https://groupietrackers.herokuapp.com/api/locations"
	RelationURL  = "https://groupietrackers.herokuapp.com/api/relation"
)

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
}

type LocationData struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
}

type LocationResponse struct {
	Index []LocationData `json:"index"`
}

type RelationData struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type RelationResponse struct {
	Index []RelationData `json:"index"`
}

// Structure to pass location data to the map
type ArtistLocation struct {
	Name  string   `json:"name"`
	Dates []string `json:"dates"`
}

type ArtistWithLocations struct {
	Artist
	Locations []ArtistLocation `json:"locations"`
}

func (a Artist) GetSpotifyURL() string {
	// Encode artist name for URL
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

func FetchArtistWithLocations(id int) (ArtistWithLocations, error) {
	// 1. Retrieve artist
	artist, err := FetchArtistByID(id)
	if err != nil {
		return ArtistWithLocations{}, err
	}

	// 2. Retrieve locations
	locResp, err := http.Get(LocationsURL)
	if err != nil {
		return ArtistWithLocations{Artist: artist}, err
	}
	defer locResp.Body.Close()

	var locationResponse LocationResponse
	if err := json.NewDecoder(locResp.Body).Decode(&locationResponse); err != nil {
		return ArtistWithLocations{Artist: artist}, err
	}

	// 3. Retrieve relations (dates per location)
	relResp, err := http.Get(RelationURL)
	if err != nil {
		return ArtistWithLocations{Artist: artist}, err
	}
	defer relResp.Body.Close()

	var relationResponse RelationResponse
	if err := json.NewDecoder(relResp.Body).Decode(&relationResponse); err != nil {
		return ArtistWithLocations{Artist: artist}, err
	}

	// 4. Find data for this artist
	var locations []ArtistLocation

	// Find artist's locations
	for _, locData := range locationResponse.Index {
		if locData.ID == id {
			// Find dates for this artist
			var datesLocations map[string][]string
			for _, relData := range relationResponse.Index {
				if relData.ID == id {
					datesLocations = relData.DatesLocations
					break
				}
			}

			// Build list of locations with their dates
			for _, locName := range locData.Locations {
				dates := datesLocations[locName]
				locations = append(locations, ArtistLocation{
					Name:  locName,
					Dates: dates,
				})
			}
			break
		}
	}

	return ArtistWithLocations{
		Artist:    artist,
		Locations: locations,
	}, nil
}

func FetchLocations() (LocationResponse, error) {
	resp, err := http.Get(LocationsURL)
	if err != nil {
		return LocationResponse{}, err
	}
	defer resp.Body.Close()

	var locationResponse LocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&locationResponse); err != nil {
		return LocationResponse{}, err
	}

	return locationResponse, nil
}

func FetchRelations() (RelationResponse, error) {
	resp, err := http.Get(RelationURL)
	if err != nil {
		return RelationResponse{}, err
	}
	defer resp.Body.Close()

	var relationResponse RelationResponse
	if err := json.NewDecoder(resp.Body).Decode(&relationResponse); err != nil {
		return RelationResponse{}, err
	}

	return relationResponse, nil
}
