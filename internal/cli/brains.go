package cli

import (
	"fmt"
	"strings"

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

func openBrain(name string) (brain.Brain, error) {
	r, err := loadRegistry()
	if err != nil {
		return brain.Brain{}, err
	}
	return r.Get(name)
}

func noBrainError(command string) error {
	var names []string
	if r, err := loadRegistry(); err == nil {
		for _, b := range r.List() {
			names = append(names, b.Name)
		}
	}
	registered := "none registered"
	if len(names) > 0 {
		registered = "registered: " + strings.Join(names, ", ")
	}
	return &Error{Code: codeBrain, Err: fmt.Errorf("%s: --brain is required (%s)", command, registered)}
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
