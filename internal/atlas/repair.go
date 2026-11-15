package atlas

import (
	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/engram"
)

// Fix repairs what can be repaired: it writes every missing inverse, then
// regenerates every existing SYNAPSES block from config. A repair that fails
// is returned and its problem stays.
func Fix(b brain.Brain, cfg config.Config, problems []Problem) []error {
	var failed []error
	for _, p := range problems {
		if p.Kind != MissingInverse {
			continue
		}
		source, target := brain.Ref{Brain: b, File: p.Engram}, brain.Ref{Brain: b, File: p.target}
		if _, err := engram.Attach(cfg, source, p.field, target); err != nil {
			failed = append(failed, err)
		}
	}
	files, err := engramFiles(b.Location)
	if err != nil {
		return append(failed, err)
	}
	for _, file := range files {
		if _, err := engram.RewriteBlock(brain.Ref{Brain: b, File: file}, cfg); err != nil {
			failed = append(failed, err)
		}
	}
	return failed
}
