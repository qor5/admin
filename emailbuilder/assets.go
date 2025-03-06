package emailbuilder

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embeddedDist embed.FS

// EmailBuilderDist return dist
var EmailBuilderDist, _ = fs.Sub(embeddedDist, "dist")
