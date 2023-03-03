package templates

import (
	"embed"
	"text/template"
)

var (
	//go:embed resources/*
	resources embed.FS
	Resources = template.Must(template.ParseFS(resources, "resources/*"))
)
