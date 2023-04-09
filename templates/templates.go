package templates

import (
	"embed"
	"text/template"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var FuncMap = template.FuncMap{
	// The name "title" is what the function will be called in the template text.
	"title": cases.Title(language.AmericanEnglish).String,
	"dayDate": func(t time.Time) string {
		return t.Format("02 Jan 2006")
	},
	"unixTs": func(t time.Time) int64 {
		return t.Unix()
	},
	"uuidStr": func(u uuid.UUID) string {
		return u.String()
	},
}

var (
	//go:embed resources/*
	resources embed.FS
	Resources = template.Must(
		template.New("any").Funcs(FuncMap).ParseFS(resources, "resources/*"),
	)
)

// func LoadTemplates() (*template.Template, error) {
// 	// Use the embedded filesystem to load the templates
// 	templateFiles, err := embedFS.ReadFile("templates/index.html")
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Parse all the template files using ParseFS
// 	t := template.New("")
// 	t, err = t.ParseFS(templateFiles, "templates/**/*.html")
// 	if err != nil {
// 		return nil, err
// 	}

// 	return t, nil
// }
