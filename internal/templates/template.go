package templates

import (
	"html/template"
	"log"
)

var Templates *template.Template

func Init() {
	var err error
	Templates, err = template.ParseGlob("templates/*.gohtml")
	if err != nil {
		log.Fatal("Erreur templates:", err)
	}
}
