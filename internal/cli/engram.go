package cli

import (
	"io"
	"time"

	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/engram"
)

func runCreateEngram(inv Invocation) (any, error) {
	b, err := openBrain(inv.Flags["brain"])
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(b.Location)
	if err != nil {
		return nil, err
	}
	body := inv.Flags["body"]
	if body == "-" {
		data, err := io.ReadAll(inv.Stdin)
		if err != nil {
			return nil, err
		}
		body = string(data)
	}
	return engram.Create(b, cfg.Engrams, engram.Options{Title: inv.Flags["title"], Body: body}, time.Now())
}
