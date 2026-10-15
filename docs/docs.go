// Package docs embeds the manual pages, <name>.<section>.md (§17.4).
package docs

import "embed"

//go:embed *.md
var FS embed.FS
