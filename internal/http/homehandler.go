package httphandlers

import (
	"log"
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/templates"
	"github.com/YajiTV/groupie-tracker/internal/util"
)

type HomeData struct {
	Title   string
	Artists []util.Artist
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	artists, err := util.FetchArtists()
	if err != nil {
		http.Error(w, "Erreur API", 500)
		log.Println(err)
		return
	}

	data := HomeData{
		Title:   "Groupie Tracker",
		Artists: artists,
	}

	templates.Templates.ExecuteTemplate(w, "index.gohtml", data)
}
