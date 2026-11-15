package cli

import (
	"fmt"

	"github.com/joakimmj/memoria/internal/atlas"
	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
)

// The write already happened, so a failed refresh only warns.
func refreshAtlas(inv Invocation, b brain.Brain, cfg config.Config) {
	if _, err := atlas.Refresh(b, cfg); err != nil && inv.Stderr != nil {
		fmt.Fprintf(inv.Stderr, "mia: warning: %v\n", err)
	}
}
