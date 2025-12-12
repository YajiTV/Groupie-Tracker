package httphandlers

import (
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/templates"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	if err := templates.Templates.ExecuteTemplate(w, "error404.gohtml", map[string]any{
		"Title": "404 — Page introuvable",
	}); err != nil {
		http.Error(w, "404 — Page introuvable", http.StatusNotFound)
	}
}
