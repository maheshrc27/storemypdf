package assets

import (
	"embed"
)

//go:embed "emails" "migrations" "static"
var EmbeddedFiles embed.FS
