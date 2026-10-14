package cub

import (
	"embed"
	_ "embed"
)

//go:embed cub_historico.csv
var CUBHistoricoCSV string

//go:embed templates/*.html templates/partials/*.html
var TemplatesFS embed.FS

//go:embed static/css/* static/js/*
var StaticFS embed.FS
