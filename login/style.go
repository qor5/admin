package login

import (
	_ "embed"
)

//go:embed style.css
var defaultStyleCSS string

var StyleCSS = defaultStyleCSS
