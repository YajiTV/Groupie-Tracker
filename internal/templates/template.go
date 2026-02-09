package templates

import (
	"html/template"
	"log"
)

var Templates *template.Template

// Init loads all templates with custom functions
func Init() {
	var err error

	// Load templates with custom functions
	Templates, err = template.New("").Funcs(TemplateFuncs()).ParseGlob("templates/*.gohtml")
	if err != nil {
		log.Fatalf("Erreur critique lors du chargement des templates: %v", err)
	}
}
