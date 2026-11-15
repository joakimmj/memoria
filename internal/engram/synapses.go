package engram

import (
	"maps"
	"slices"
	"strings"

	"github.com/joakimmj/memoria/internal/config"
)

const (
	synapsesSection = "SYNAPSES"
	blockEnd        = "***"
)

// formatBlock returns the SYNAPSES block body for values, which maps a field
// to its target filenames; nil when there is no field to show.
func formatBlock(fields map[string]config.Synapse, values map[string][]string) []string {
	labels := slices.Concat(slices.Collect(maps.Keys(fields)), slices.Collect(maps.Keys(values)))
	slices.Sort(labels)
	var items []string
	for _, l := range slices.Compact(labels) {
		f, configured := fields[l]
		targets := values[l]
		if configured && !f.ShowEmpty && len(targets) == 0 {
			continue
		}
		items = append(items, fieldItem(l, targets))
	}
	if items == nil {
		return nil
	}
	return append(items, blockEnd)
}

func fieldItem(label string, targets []string) string {
	item := "- **" + label + ":**"
	if len(targets) == 0 {
		return item
	}
	links := make([]string, len(targets))
	for i, t := range targets {
		links[i] = link(t)
	}
	return item + " " + strings.Join(links, ", ")
}

func link(file string) string {
	return "[" + strings.TrimSuffix(file, ".md") + "](" + file + ")"
}
