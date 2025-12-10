package app

import (
	"net/http"

	httphandlers "github.com/YajiTV/groupie-tracker/internal/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", httphandlers.HomeHandler)
	mux.HandleFunc("/artist/", httphandlers.ArtistHandler)

	return mux
}
