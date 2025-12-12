package app

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	httphandlers "github.com/YajiTV/groupie-tracker/internal/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Static
	fs := http.FileServer(http.Dir(StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			httphandlers.NotFoundHandler(w, r)
			return
		}
		httphandlers.HomeHandler(w, r)
	})

	mux.HandleFunc("/artist/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/artist/")
		rest = strings.Trim(rest, "/")

		if rest == "" || strings.Contains(rest, "/") {
			httphandlers.NotFoundHandler(w, r)
			return
		}

		id, err := strconv.Atoi(rest)
		if err != nil || id <= 0 {
			httphandlers.NotFoundHandler(w, r)
			return
		}
		httphandlers.ArtistHandler(w, r)
	})

	return mux
}

func Start() {
	log.Printf("Serveur sur http://localhost%s\n", Port)
	log.Fatal(http.ListenAndServe(Port, SetupRouter()))
}
