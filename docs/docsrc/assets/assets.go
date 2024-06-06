package assets

import "embed"

//go:embed **.*
var Assets embed.FS

//go:embed favicon.ico
var Favicon []byte
