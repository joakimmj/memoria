package cli

import (
	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/engram"
)

func runAttachSynapse(inv Invocation) (any, error) {
	r, err := loadRegistry()
	if err != nil {
		return nil, err
	}
	source, target, err := r.ResolvePair(inv.Args[0], inv.Args[2], inv.Flags["brain"])
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(source.Brain.Location)
	if err != nil {
		return nil, err
	}
	return engram.Attach(cfg, source, inv.Args[1], target)
}
