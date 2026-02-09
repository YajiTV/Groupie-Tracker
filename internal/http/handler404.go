package httphandlers

import (
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/templates"
)

type Error404Data struct {
	Title  string
	Path   string
	Method string
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	data := Error404Data{
		Title:  "404 — Page introuvable",
		Path:   r.URL.Path,
		Method: r.Method,
	}

	if err := templates.Templates.ExecuteTemplate(w, "error404.gohtml", data); err != nil {
		http.Error(w, "404 — Page introuvable", http.StatusNotFound)
	}
}
