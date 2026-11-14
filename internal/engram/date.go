package engram

import (
	"strings"
	"time"
)

// Longest first, so YYYY wins over YY.
var dateTokens = []struct{ token, layout string }{
	{"YYYY", "2006"},
	{"YY", "06"},
	{"MM", "01"},
	{"DD", "02"},
	{"HH", "15"},
	{"mm", "04"},
	{"ss", "05"},
}

// formatDate writes t with format's tokens: YYYY, YY, MM, DD, HH, mm, ss.
// Anything else is written as is.
func formatDate(t time.Time, format string) string {
	var b strings.Builder
next:
	for i := 0; i < len(format); {
		for _, dt := range dateTokens {
			if strings.HasPrefix(format[i:], dt.token) {
				b.WriteString(t.Format(dt.layout))
				i += len(dt.token)
				continue next
			}
		}
		b.WriteByte(format[i])
		i++
	}
	return b.String()
}
