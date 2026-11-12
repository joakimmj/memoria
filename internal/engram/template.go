package engram

import "fmt"

var errManyBodies = fmt.Errorf("more than one %%body%%: %w", ErrTemplate)

type vars struct {
	title, date string
}

// render expands a content template. body goes where %body% is, or at the
// end without one; the position answered is where body starts.
//
// TODO(human): implement. Rules (template_test.go holds the cases):
//   - %title% and %date% expand to v.title and v.date, everywhere
//   - %body% is removed; body is written in its place
//   - no %body% → body goes at the end
//   - more than one %body% → errManyBodies
//   - the position is where body starts: Line from 1, Col in bytes from 0,
//     counted in the expanded text
//   - a title holding "%date%" is written as given, never expanded again
//
// Hints:
//   - strings.Count, then strings.Cut(template, "%body%") splits it in two.
//   - strings.NewReplacer("%title%", ..., "%date%", ...) expands in one
//     pass, so text it wrote is not scanned again.
//   - Line: strings.Count(before, "\n") + 1. Col: bytes after the last
//     "\n" (strings.LastIndexByte returns -1 when there is none).
func render(template string, v vars, body string) (string, Position, error) {
	return template, Position{}, nil
}
