package engram

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/markdown"
)

// Section is the fenced section holding an engram's synapses.
const Section = "SYNAPSES"

const blockEnd = "***"

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

// A marker such as "[ ]" before the label is allowed and dropped.
var fieldLine = regexp.MustCompile(`^- (?:\[[^\]]*\]\s+)?\*\*([A-Za-z0-9_]+):\*\*(.*)$`)

// parseBlock ignores lines that are not field items, and anything in an item
// that is not a link to an engram.
func parseBlock(body []string) map[string][]string {
	values := map[string][]string{}
	for _, line := range body {
		m := fieldLine.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		label, targets := m[1], values[m[1]]
		if targets == nil {
			targets = []string{}
		}
		for _, l := range markdown.Links(m[2]) {
			file, ok := markdown.EngramTarget(l.Dest)
			if ok && !slices.Contains(targets, file) {
				targets = append(targets, file)
			}
		}
		values[label] = targets
	}
	return values
}

// Synapses returns the targets of each field in an engram's SYNAPSES block.
func Synapses(lines []string) map[string][]string {
	body, _ := markdown.Section(lines, Section)
	return parseBlock(body)
}

// writeSynapses regenerates the block, under the frontmatter when new.
func writeSynapses(lines []string, fields map[string]config.Synapse, values map[string][]string) []string {
	block := formatBlock(fields, values)
	if block == nil {
		return lines
	}
	if _, ok := markdown.Section(lines, Section); ok {
		return markdown.SetSection(lines, Section, block, 0)
	}
	at := markdown.FrontmatterEnd(lines)
	if at < len(lines) && lines[at] != "" {
		lines = slices.Concat(lines[:at], []string{""}, lines[at:])
	}
	return markdown.SetSection(lines, Section, block, at)
}

// Body returns lines without the frontmatter, the SYNAPSES block and the
// blank line that follows a header.
func Body(lines []string) []string {
	start := markdown.FrontmatterEnd(lines)
	rest := lines[start:]
	if open, close, ok := markdown.SectionBounds(rest, Section); ok {
		if open == 0 {
			start += close + 1
		} else {
			return slices.Concat(rest[:open], rest[close+1:])
		}
	}
	if start > 0 && start < len(lines) && lines[start] == "" {
		start++
	}
	return slices.Clone(lines[start:])
}

// FieldLine returns the index of field's item in the SYNAPSES block, or -1.
func FieldLine(lines []string, field string) int {
	open, close, ok := markdown.SectionBounds(lines, Section)
	if !ok {
		return -1
	}
	for i := open + 1; i < close; i++ {
		if m := fieldLine.FindStringSubmatch(strings.TrimSpace(lines[i])); m != nil && m[1] == field {
			return i
		}
	}
	return -1
}
