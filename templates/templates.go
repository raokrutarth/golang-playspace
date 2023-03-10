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
