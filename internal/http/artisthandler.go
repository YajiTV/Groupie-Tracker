package httphandlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

type ArtistData struct {
	Artist util.ArtistWithLocations
}

func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/artist/"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	artistWithLocations, err := util.FetchArtistWithLocations(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := ArtistData{Artist: artistWithLocations}
	templates.Templates.ExecuteTemplate(w, "artist.gohtml", data)
}
