package emailbuilder

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embeddedDist embed.FS

// EmailBuilderDist 返回 dist 目录的文件系统
var EmailBuilderDist, _ = fs.Sub(embeddedDist, "dist")
