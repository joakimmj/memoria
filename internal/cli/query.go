package cli

import (
	"fmt"

	"github.com/joakimmj/memoria/internal/atlas"
	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/query"
)

// brainsFor returns the brains --brain names, or every brain when it names
// none; named says which.
func brainsFor(inv Invocation) (brains []brain.Brain, named bool, err error) {
	r, err := loadRegistry()
	if err != nil {
		return nil, false, err
	}
	names := inv.Repeated["brain"]
	if len(names) == 0 {
		return r.List(), false, nil
	}
	for _, name := range names {
		b, err := r.Get(name)
		if err != nil {
			return nil, true, err
		}
		brains = append(brains, b)
	}
	return brains, true, nil
}

// A brain that fails is skipped with a warning, unless named: one missing
// folder must not hide every other brain.
func sources(inv Invocation, brains []brain.Brain, named bool) ([]query.Source, error) {
	var out []query.Source
	for _, b := range brains {
		s, err := source(b)
		if err != nil && named {
			return nil, err
		}
		if err != nil {
			fmt.Fprintf(inv.Stderr, "mia: warning: skipping %v\n", err)
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func source(b brain.Brain) (query.Source, error) {
	if !b.Exists() {
		return query.Source{}, fmt.Errorf("(%s) missing folder %s", b.Name, b.Location)
	}
	cfg, err := config.Load(b.Location)
	if err != nil {
		return query.Source{}, fmt.Errorf("(%s) %w", b.Name, err)
	}
	a, err := atlas.Refresh(b, cfg)
	if err != nil {
		return query.Source{}, err
	}
	return query.Source{Brain: b, Atlas: a}, nil
}

func runEngrams(inv Invocation) (any, error) {
	brains, named, err := brainsFor(inv)
	if err != nil {
		return nil, err
	}
	srcs, err := sources(inv, brains, named)
	if err != nil {
		return nil, err
	}
	return query.Engrams(srcs), nil
}

func runEngram(inv Invocation) (any, error) {
	r, err := loadRegistry()
	if err != nil {
		return nil, err
	}
	ref, err := r.Resolve(inv.Args[0], inv.Flags["brain"])
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(ref.Brain.Location)
	if err != nil {
		return nil, err
	}
	a, err := atlas.Refresh(ref.Brain, cfg)
	if err != nil {
		return nil, err
	}
	return query.Engram(ref, a, cfg.Engrams, inv.Flags["body"] == "true")
}

func runTasks(inv Invocation) (any, error) {
	state := inv.Flags["state"]
	if state != "" && state != config.NotDone && state != config.Done {
		return nil, usageErrorf("tasks: --state %q: not_done or done", state)
	}
	filter := query.Filter{State: state}
	arg, byEngram := inv.Flags["engram"]
	if !byEngram {
		brains, named, err := brainsFor(inv)
		if err != nil {
			return nil, err
		}
		srcs, err := sources(inv, brains, named)
		if err != nil {
			return nil, err
		}
		return query.Tasks(srcs, filter), nil
	}

	brainNames := inv.Repeated["brain"]
	if len(brainNames) > 1 {
		return nil, usageErrorf("tasks: --engram takes one --brain at most")
	}
	r, err := loadRegistry()
	if err != nil {
		return nil, err
	}
	ref, err := r.Resolve(arg, firstOr(brainNames, ""))
	if err != nil {
		return nil, err
	}
	s, err := source(ref.Brain)
	if err != nil {
		return nil, err
	}
	if _, ok := s.Atlas.Engrams[ref.File]; !ok {
		return nil, fmt.Errorf("(%s) %s: %w", ref.Brain.Name, ref.File, brain.ErrNoEngram)
	}
	filter.Engram = ref.File
	return query.Tasks([]query.Source{s}, filter), nil
}

func firstOr(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}
