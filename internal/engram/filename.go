package engram

import (
	"fmt"
	"time"

	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/slug"
)

func filename(cfg config.Filename, title string, now time.Time) (string, error) {
	switch cfg.Prefix {
	case "date":
		date := formatDate(now, cfg.DateFormat)
		if title == "" {
			return date + ".md", nil
		}
		s, err := slugOf(title)
		if err != nil {
			return "", err
		}
		return date + "_" + s + ".md", nil
	case "none":
		if title == "" {
			return "", ErrNoTitle
		}
		s, err := slugOf(title)
		if err != nil {
			return "", err
		}
		return s + ".md", nil
	default:
		return "", fmt.Errorf("%q: %w", cfg.Prefix, ErrPrefix)
	}
}

func slugOf(title string) (string, error) {
	s := slug.Make(title)
	if s == "" {
		return "", fmt.Errorf("%q: %w", title, ErrEmptySlug)
	}
	return s, nil
}
