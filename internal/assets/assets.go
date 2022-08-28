package assets

import (
	"embed"
)

var (
	//go:embed *.tmpl
	FS embed.FS
)
