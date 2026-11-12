package engram

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, time.August, 1, 9, 5, 3, 0, time.Local)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		format, want string
	}{
		{"YYYY-MM-DD", "2026-08-01"},
		{"DD.MM.YYYY", "01.08.2026"},
		{"YY", "26"},
		{"HH:mm:ss", "09:05:03"},
		{"YYYYMMDD_HHmm", "20260801_0905"},
		{"YYYYY", "2026Y"},
		{"MMM", "08M"},
		{"week of YYYY", "week of 2026"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := formatDate(testNow, tt.format); got != tt.want {
			t.Errorf("formatDate(%q) = %q, want %q", tt.format, got, tt.want)
		}
	}
}
