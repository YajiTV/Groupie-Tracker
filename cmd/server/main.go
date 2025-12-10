package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/YajiTV/groupie-tracker/internal/app"
	"github.com/YajiTV/groupie-tracker/internal/templates"
)

func main() {
	templates.Init()
	router := app.SetupRouter()

	port := ":8080"
	fmt.Printf("Serveur sur http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
