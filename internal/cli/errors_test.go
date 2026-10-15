package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"
)

func TestError(t *testing.T) {
	base := &Error{Code: "collision", Err: fs.ErrExist}
	tests := []struct {
		name     string
		err      error
		wantMsg  string
		wantCode string
		wantIs   error // nil: nothing to check
	}{
		{
			name:     "coded error",
			err:      base,
			wantMsg:  "file already exists",
			wantCode: "collision",
			wantIs:   fs.ErrExist,
		},
		{
			name:     "wrapped coded error",
			err:      fmt.Errorf("(work) x.md: %w", base),
			wantMsg:  "(work) x.md: file already exists",
			wantCode: "collision",
			wantIs:   fs.ErrExist,
		},
		{
			name:     "plain error has no code",
			err:      errors.New("boom"),
			wantMsg:  "boom",
			wantCode: "",
		},
		{
			name:     "usageErrorf",
			err:      usageErrorf("unknown command %q", "nope"),
			wantMsg:  `unknown command "nope"`,
			wantCode: codeUsage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
			if got := codeOf(tt.err); got != tt.wantCode {
				t.Errorf("codeOf() = %q, want %q", got, tt.wantCode)
			}
			if tt.wantIs != nil && !errors.Is(tt.err, tt.wantIs) {
				t.Errorf("errors.Is(err, %v) = false, want true", tt.wantIs)
			}
		})
	}
}
