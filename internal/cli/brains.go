package cli

import (
	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
)

type brainStatus struct {
	brain.Brain
	Exists bool `json:"exists"`
	HasDNA bool `json:"has_dna"`
}

func loadRegistry() (*brain.Registry, error) {
	paths, err := config.Locations()
	if err != nil {
		return nil, err
	}
	return brain.Load(paths.Registry)
}

func runBrains(Invocation) (any, error) {
	r, err := loadRegistry()
	if err != nil {
		return nil, err
	}
	out := []brainStatus{}
	for _, b := range r.List() {
		out = append(out, brainStatus{Brain: b, Exists: b.Exists(), HasDNA: b.HasDNA()})
	}
	return out, nil
}

func runRegisterBrain(inv Invocation) (any, error) {
	r, err := loadRegistry()
	if err != nil {
		return nil, err
	}
	b, err := r.Register(inv.Args[0], inv.Flags["name"])
	if err != nil {
		return nil, err
	}
	return b, r.Save()
}

func runDeregisterBrain(inv Invocation) (any, error) {
	r, err := loadRegistry()
	if err != nil {
		return nil, err
	}
	b, err := r.Deregister(inv.Args[0])
	if err != nil {
		return nil, err
	}
	return b, r.Save()
}
