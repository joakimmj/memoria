package cli

import (
	"fmt"

	"github.com/joakimmj/memoria/internal/atlas"
	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
)

func runCheck(inv Invocation) (any, error) {
	brains, named, err := brainsFor(inv)
	if err != nil {
		return nil, err
	}
	problems := []atlas.Problem{}
	for _, b := range brains {
		found, err := rebuilt(b)
		if err != nil && named {
			return nil, err
		}
		if err != nil {
			fmt.Fprintf(inv.Stderr, "mia: warning: skipping %v\n", err)
			continue
		}
		problems = append(problems, found...)
	}
	if inv.Flags["strict"] == "true" && len(problems) > 0 {
		return nil, &Error{Code: codeProblems, Err: fmt.Errorf("%d problems", len(problems)), Result: problems}
	}
	return problems, nil
}

func runRebuild(inv Invocation) (any, error) {
	b, err := openBrain(inv.Flags["brain"])
	if err != nil {
		return nil, err
	}
	problems, err := rebuilt(b)
	if err != nil || inv.Flags["fix"] != "true" {
		return problems, err
	}
	cfg, err := config.Load(b.Location)
	if err != nil {
		return nil, err
	}
	for _, failed := range atlas.Fix(b, cfg, problems) {
		fmt.Fprintf(inv.Stderr, "mia: warning: repair failed: %v\n", failed)
	}
	return rebuilt(b)
}

func rebuilt(b brain.Brain) ([]atlas.Problem, error) {
	if !b.Exists() {
		return nil, fmt.Errorf("(%s) missing folder %s", b.Name, b.Location)
	}
	cfg, err := config.Load(b.Location)
	if err != nil {
		return nil, fmt.Errorf("(%s) %w", b.Name, err)
	}
	a, err := atlas.Rebuild(b, cfg)
	if err != nil {
		return nil, err
	}
	return atlas.Check(b, a, cfg), nil
}
