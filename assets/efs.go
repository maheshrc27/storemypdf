package assets

import (
	"embed"
)

//go:embed "emails" "static"
var EmbeddedFiles embed.FS
