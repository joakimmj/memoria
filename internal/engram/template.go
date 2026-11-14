package engram

import (
	"fmt"
	"strings"
)

var errManyBodies = fmt.Errorf("%w: more than one %%body%%", ErrTemplate)

type vars struct {
	title, date string
}

// render expands a content template. body goes where %body% is, or at the
// end without one; the position answered is where body starts.
func render(template string, v vars, body string) (string, Position, error) {
	if strings.Count(template, "%body%") > 1 {
		return "", Position{}, errManyBodies
	}
	expand := strings.NewReplacer("%title%", v.title, "%date%", v.date).Replace
	before, after, _ := strings.Cut(template, "%body%")
	before, after = expand(before), expand(after)
	return before + body + after, endOf(before), nil
}

func endOf(text string) Position {
	return Position{
		Line: strings.Count(text, "\n") + 1,
		Col:  len(text) - strings.LastIndexByte(text, '\n') - 1,
	}
}
