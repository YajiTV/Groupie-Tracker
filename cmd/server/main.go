package main

import (
	"github.com/YajiTV/groupie-tracker/internal/app"
	"github.com/YajiTV/groupie-tracker/internal/templates"
)

func main() {
	templates.Init()
	app.Start()
}
