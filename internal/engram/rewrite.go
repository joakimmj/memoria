package engram

import (
	"fmt"
	"os"
	"slices"

	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/markdown"
)

// RewriteBlock regenerates ref's SYNAPSES block from config, and reports
// whether the file changed. An engram without a block is left alone.
func RewriteBlock(ref brain.Ref, cfg config.Config) (bool, error) {
	data, err := os.ReadFile(ref.Path())
	if err != nil {
		return false, fmt.Errorf("(%s) %w", ref.Brain.Name, err)
	}
	lines := markdown.Lines(string(data))
	if _, _, ok := markdown.SectionBounds(lines, Section); !ok {
		return false, nil
	}
	written := writeSynapses(lines, cfg.Synapses, Synapses(lines))
	if slices.Equal(written, lines) {
		return false, nil
	}
	if err := replaceFile(ref.Path(), markdown.Join(written)); err != nil {
		return false, fmt.Errorf("(%s) %w", ref.Brain.Name, err)
	}
	return true, nil
}
