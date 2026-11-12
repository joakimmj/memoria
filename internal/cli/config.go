package cli

import "github.com/joakimmj/memoria/internal/config"

func runConfig(inv Invocation) (any, error) {
	name, withBrain := inv.Flags["brain"]
	if inv.Flags["defaults"] == "true" {
		if withBrain {
			return nil, usageErrorf("config: --defaults and --brain contradict each other")
		}
		return config.Defaults()
	}
	if !withBrain {
		return config.Load("")
	}
	r, err := loadRegistry()
	if err != nil {
		return nil, err
	}
	b, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	return config.Load(b.Location)
}
