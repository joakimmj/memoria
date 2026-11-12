package cli

import "github.com/joakimmj/memoria/internal/config"

func runConfig(inv Invocation) (any, error) {
	if inv.Flags["defaults"] == "true" {
		return config.Defaults()
	}
	return config.Load("")
}
