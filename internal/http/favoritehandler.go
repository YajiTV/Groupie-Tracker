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
	// Auth via cookie session (ton système actuel)
	session, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// URL attendue : /favorite/toggle/12
	idStr := strings.TrimPrefix(r.URL.Path, "/favorite/toggle/")
	artistID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID d'artiste invalide", http.StatusBadRequest)
		return
	}

	// Récupérer l'artiste via l'API (fonction déjà existante)
	artist, err := util.FetchArtistByID(artistID)
	if err != nil {
		http.Error(w, "Artiste introuvable", http.StatusNotFound)
		return
	}

	// Toggle : si déjà favori => remove, sinon add
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

	// Retour à la page précédente (artist page)
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/"
	}
	http.Redirect(w, r, ref, http.StatusSeeOther)
}
