package templates

import "embed"

//go:embed all:layouts all:movies *.html
var Files embed.FS

