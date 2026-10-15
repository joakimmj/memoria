// Package docs embeds the manual pages, one <name>.<section>.md file each.
package docs

import "embed"

//go:embed *.md
var FS embed.FS
