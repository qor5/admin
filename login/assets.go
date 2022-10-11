package login

import (
	"embed"
)

//go:embed assets
var assetsFS embed.FS

var assetsPathPrefix = "/auth/assets/"
var (
	styleCSSURL = assetsPathPrefix + "style.css"
	zxcvbnJSURL = assetsPathPrefix + "zxcvbn.js"
)
