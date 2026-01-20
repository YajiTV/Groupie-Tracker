package storage

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Favorite struct {
	UserID      int       `json:"user_id"`
	ArtistID    int       `json:"artist_id"`
	ArtistName  string    `json:"artist_name"`
	ArtistImage string    `json:"artist_image"`
	AddedAt     time.Time `json:"added_at"`
}

type favoritesData struct {
	Favorites []Favorite `json:"favorites"`
}

var (
	favFile  = "data/favorites.json"
	favMutex sync.RWMutex
)

func InitFavorites() error {
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}
	if _, err := os.Stat(favFile); os.IsNotExist(err) {
		return saveFav(favoritesData{Favorites: []Favorite{}})
	}
	return nil
}

func GetFavorites(userID int) ([]Favorite, error) {
	favMutex.RLock()
	defer favMutex.RUnlock()

	data, err := loadFav()
	if err != nil {
		return nil, err
	}

	out := []Favorite{}
	for _, f := range data.Favorites {
		if f.UserID == userID {
			out = append(out, f)
		}
	}
	return out, nil
}

func IsFavorite(userID, artistID int) (bool, error) {
	favs, err := GetFavorites(userID)
	if err != nil {
		return false, err
	}
	for _, f := range favs {
		if f.ArtistID == artistID {
			return true, nil
		}
	}
	return false, nil
}

func AddFavorite(f Favorite) error {
	favMutex.Lock()
	defer favMutex.Unlock()

	data, err := loadFav()
	if err != nil {
		return err
	}

	for _, existing := range data.Favorites {
		if existing.UserID == f.UserID && existing.ArtistID == f.ArtistID {
			return nil // déjà là
		}
	}

	data.Favorites = append(data.Favorites, f)
	return saveFav(data)
}

func RemoveFavorite(userID, artistID int) error {
	favMutex.Lock()
	defer favMutex.Unlock()

	data, err := loadFav()
	if err != nil {
		return err
	}

	out := data.Favorites[:0]
	for _, f := range data.Favorites {
		if !(f.UserID == userID && f.ArtistID == artistID) {
			out = append(out, f)
		}
	}
	data.Favorites = out
	return saveFav(data)
}

func loadFav() (favoritesData, error) {
	file, err := os.Open(favFile)
	if err != nil {
		return favoritesData{}, err
	}
	defer file.Close()

	var data favoritesData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return favoritesData{}, err
	}
	return data, nil
}

func saveFav(data favoritesData) error {
	file, err := os.Create(favFile)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
