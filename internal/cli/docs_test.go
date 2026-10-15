package cli

import (
	"io/fs"
	"testing"

	"github.com/joakimmj/memoria/docs"
)

// TestEveryCommandHasPage: no command without a manual page.
func TestEveryCommandHasPage(t *testing.T) {
	for _, cmd := range table {
		t.Run(cmd.Name, func(t *testing.T) {
			name := "mia-" + cmd.Name + ".1.md"
			if _, err := fs.Stat(docs.FS, name); err != nil {
				t.Errorf("command %q has no manual page docs/%s: %v", cmd.Name, name, err)
			}
		})
	}
}
