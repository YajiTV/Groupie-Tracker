package httphandlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/YajiTV/groupie-tracker/internal/auth"
	"github.com/YajiTV/groupie-tracker/internal/storage"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

func ToggleFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	// Auth via cookie session (current system)
	session, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Expected URL: /favorite/toggle/12
	idStr := strings.TrimPrefix(r.URL.Path, "/favorite/toggle/")
	artistID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID d'artiste invalide", http.StatusBadRequest)
		return
	}

	// Retrieve artist via API (existing function)
	artist, err := util.FetchArtistByID(artistID)
	if err != nil {
		http.Error(w, "Artiste introuvable", http.StatusNotFound)
		return
	}

	// Toggle: if already favorite => remove, otherwise add
	isFav, _ := storage.IsFavorite(session.UserID, artistID)
	if isFav {
		_ = storage.RemoveFavorite(session.UserID, artistID)
	} else {
		_ = storage.AddFavorite(storage.Favorite{
			UserID:      session.UserID,
			ArtistID:    artistID,
			ArtistName:  artist.Name,
			ArtistImage: artist.Image,
			AddedAt:     time.Now(),
		})
	}

	// Return to previous page (artist page)
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/"
	}
	http.Redirect(w, r, ref, http.StatusSeeOther)
}
