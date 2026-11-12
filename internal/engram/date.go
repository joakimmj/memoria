package engram

import "time"

// formatDate writes t with format's tokens: YYYY, YY, MM, DD, HH, mm, ss.
// Anything else is written as is.
//
// TODO(human): implement. Rules (date_test.go holds the cases):
//   - scan format left to right; at each position, the longest token that
//     matches wins, so "YYYY" is one year, not "YY" twice
//   - a token writes its zero-padded value: YYYY 2026, YY 26, MM 08, DD 01,
//     HH 09 (24-hour), mm 05 (minutes), ss 03
//   - anything else is copied as is
//
// Hints:
//   - Go's own t.Format uses a reference date ("2006-01-02"), not letters;
//     t.Format("YYYY") writes the text "YYYY". Use it per token
//     (t.Format("2006")), or fmt.Sprintf("%02d", t.Month()).
//   - strings.HasPrefix(format[i:], tok) tests a token at position i. Try
//     the 4-letter token before the 2-letter ones.
//   - strings.Builder collects the result.
func formatDate(t time.Time, format string) string {
	return format
}
