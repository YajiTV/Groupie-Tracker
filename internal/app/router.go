package app

import (
	"log"
	"net/http"

	httphandlers "github.com/YajiTV/groupie-tracker/internal/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", httphandlers.HomeHandler)
	mux.HandleFunc("/artist/", httphandlers.ArtistHandler)

	return mux
}

func Start() {
	log.Printf("Serveur sur http://localhost%s\n", Port)
	log.Fatal(http.ListenAndServe(Port, SetupRouter()))
}
