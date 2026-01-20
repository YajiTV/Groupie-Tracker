package templates

import (
	"html/template"
	"log"
)

var Templates *template.Template

// Init charge tous les templates avec les fonctions personnalisées
func Init() {
	var err error

	// Charger les templates avec les fonctions personnalisées
	Templates, err = template.New("").Funcs(TemplateFuncs()).ParseGlob("templates/*.gohtml")
	if err != nil {
		log.Fatalf("Erreur critique lors du chargement des templates: %v", err)
	}
}
