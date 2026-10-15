package cli

import (
	"fmt"
	"slices"
	"strings"
)

// Page markdown subset: `# title` (ignored), `## HEADING`, paragraphs,
// `- item` (indented lines continue), fenced code, **bold** and `code`.

// sectionOrder is the allowed sections, in display order.
var sectionOrder = []string{
	"NAME", "SYNOPSIS", "DESCRIPTION", "COMMANDS", "OPTIONS", "OUTPUT",
	"ERRORS", "EXIT STATUS", "EXAMPLES", "SEE ALSO",
}

// generatedSections come from the command table, not the markdown.
var generatedSections = []string{"SYNOPSIS", "COMMANDS", "OPTIONS"}

type page struct {
	name     string // "mia-help"
	section  string // "1"
	sections []pageSection
}

type pageSection struct {
	heading string
	blocks  []block
}

type blockKind int

const (
	paragraphBlock blockKind = iota
	listBlock
	codeBlock
	definitionBlock // term + description
)

type block struct {
	kind  blockKind
	term  string // definitionBlock only
	text  string
	lines []string // codeBlock only
}

// parsePage parses a page; filename gives name and section.
func parsePage(filename string, src string) (page, error) {
	base := strings.TrimSuffix(filename, ".md")
	dot := strings.LastIndexByte(base, '.')
	if dot < 0 {
		return page{}, fmt.Errorf("%s: name is not <name>.<section>.md", filename)
	}
	p := page{name: base[:dot], section: base[dot+1:]}

	var cur *pageSection
	var para []string
	inCode := false
	flush := func() {
		if len(para) > 0 && cur != nil {
			cur.blocks = append(cur.blocks, block{kind: paragraphBlock, text: strings.Join(para, " ")})
		}
		para = nil
	}

	for i, line := range strings.Split(src, "\n") {
		where := fmt.Sprintf("%s:%d", filename, i+1)
		switch {
		case inCode:
			if strings.TrimSpace(line) == "```" {
				inCode = false
				continue
			}
			last := &cur.blocks[len(cur.blocks)-1]
			last.lines = append(last.lines, line)
		case strings.HasPrefix(line, "# "):
			// the title line
		case strings.HasPrefix(line, "## "):
			flush()
			heading := strings.TrimSpace(line[3:])
			if !slices.Contains(sectionOrder, heading) {
				return page{}, fmt.Errorf("%s: unknown section %q", where, heading)
			}
			if slices.Contains(generatedSections, heading) {
				return page{}, fmt.Errorf("%s: section %q is generated from the command table", where, heading)
			}
			p.sections = append(p.sections, pageSection{heading: heading})
			cur = &p.sections[len(p.sections)-1]
		case strings.TrimSpace(line) == "":
			flush()
		case cur == nil:
			return page{}, fmt.Errorf("%s: text before the first section", where)
		case strings.TrimSpace(line) == "```":
			flush()
			inCode = true
			cur.blocks = append(cur.blocks, block{kind: codeBlock})
		case strings.HasPrefix(line, "- "):
			flush()
			cur.blocks = append(cur.blocks, block{kind: listBlock, text: strings.TrimSpace(line[2:])})
		case strings.HasPrefix(line, "  ") && len(para) == 0 && len(cur.blocks) > 0 &&
			cur.blocks[len(cur.blocks)-1].kind == listBlock:
			last := &cur.blocks[len(cur.blocks)-1]
			last.text += " " + strings.TrimSpace(line)
		default:
			para = append(para, strings.TrimSpace(line))
		}
	}
	if inCode {
		return page{}, fmt.Errorf("%s: unclosed code block", filename)
	}
	flush()
	return p, nil
}

// addGenerated adds the table-derived sections and sorts by sectionOrder.
func addGenerated(p *page) {
	switch {
	case p.name == "mia" && p.section == "1":
		p.sections = append(p.sections,
			pageSection{heading: "SYNOPSIS", blocks: []block{{kind: codeBlock, lines: []string{
				"mia <command> [args]",
				"mia [command] --help",
			}}}},
			pageSection{heading: "COMMANDS", blocks: commandList()},
		)
	case p.section == "1" && strings.HasPrefix(p.name, "mia-"):
		cmd, ok := lookup(strings.TrimPrefix(p.name, "mia-"))
		if !ok {
			break // e.g. mia-view(1)
		}
		p.sections = append(p.sections, pageSection{heading: "SYNOPSIS", blocks: []block{{
			kind: codeBlock, lines: []string{cmd.synopsis()},
		}}})
		if opts := options(cmd); len(opts) > 0 {
			p.sections = append(p.sections, pageSection{heading: "OPTIONS", blocks: opts})
		}
	}
	slices.SortStableFunc(p.sections, func(a, b pageSection) int {
		return slices.Index(sectionOrder, a.heading) - slices.Index(sectionOrder, b.heading)
	})
}

func commandList() []block {
	var bs []block
	for _, c := range table {
		bs = append(bs, block{kind: definitionBlock, term: strings.TrimPrefix(c.synopsis(), "mia "), text: c.Summary})
	}
	return bs
}

func options(cmd Command) []block {
	var bs []block
	for _, a := range cmd.Args {
		bs = append(bs, block{kind: definitionBlock, term: "<" + a.Name + ">", text: a.Summary})
	}
	for _, f := range cmd.Flags {
		text := f.Summary
		if f.Required {
			text += " (required)"
		}
		bs = append(bs, block{kind: definitionBlock, term: f.usage(false), text: text})
	}
	return bs
}

// Terminal text layout.

const (
	textIndent = 7
	textWidth  = 80
)

func renderText(p page) string {
	var b strings.Builder
	pad := strings.Repeat(" ", textIndent)
	for i, s := range p.sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(s.heading + "\n")
		for j, bl := range s.blocks {
			if j > 0 && !(bl.kind == listBlock && s.blocks[j-1].kind == listBlock) {
				b.WriteString("\n")
			}
			switch bl.kind {
			case paragraphBlock:
				writeWrapped(&b, pad, pad, plainInline(bl.text))
			case listBlock:
				writeWrapped(&b, pad+"- ", pad+"  ", plainInline(bl.text))
			case codeBlock:
				for _, l := range bl.lines {
					b.WriteString(strings.TrimRight(pad+l, " ") + "\n")
				}
			case definitionBlock:
				b.WriteString(pad + plainInline(bl.term) + "\n")
				writeWrapped(&b, pad+"    ", pad+"    ", plainInline(bl.text))
			}
		}
	}
	return b.String()
}

// writeWrapped word-wraps text, prefixing the first line and the rest.
func writeWrapped(b *strings.Builder, first, rest, text string) {
	line := first
	empty := true
	for _, w := range strings.Fields(text) {
		if !empty && len(line)+1+len(w) > textWidth {
			b.WriteString(line + "\n")
			line, empty = rest, true
		}
		if !empty {
			line += " "
		}
		line += w
		empty = false
	}
	b.WriteString(line + "\n")
}

// renderRoff renders p with the man(7) macros.
func renderRoff(p page) string {
	var b strings.Builder
	fmt.Fprintf(&b, ".TH %q %q \"\" \"mia\" \"memoria manual\"\n", strings.ToUpper(p.name), p.section)
	for _, s := range p.sections {
		fmt.Fprintf(&b, ".SH %q\n", s.heading)
		for j, bl := range s.blocks {
			switch bl.kind {
			case paragraphBlock:
				if j > 0 {
					b.WriteString(".PP\n")
				}
				b.WriteString(roffLine(roffInline(bl.text)) + "\n")
			case listBlock:
				b.WriteString(".IP \\(bu 2\n" + roffLine(roffInline(bl.text)) + "\n")
			case codeBlock:
				b.WriteString(".PP\n.RS 4\n.nf\n")
				for _, l := range bl.lines {
					b.WriteString(roffLine(roffEscape(l)) + "\n")
				}
				b.WriteString(".fi\n.RE\n")
			case definitionBlock:
				b.WriteString(".TP\n" + roffLine(`\fB`+roffEscape(bl.term)+`\fR`) + "\n")
				b.WriteString(roffLine(roffInline(bl.text)) + "\n")
			}
		}
	}
	return b.String()
}

// roffEscape escapes \ and turns - into a minus.
func roffEscape(s string) string {
	return strings.NewReplacer(`\`, `\e`, "-", `\-`).Replace(s)
}

// roffLine guards a line that would start a roff request.
func roffLine(s string) string {
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "'") {
		return `\&` + s
	}
	return s
}

func plainInline(s string) string {
	return inline(s, "", "", func(t string) string { return t })
}

func roffInline(s string) string {
	return inline(s, `\fB`, `\fR`, roffEscape)
}

// inline wraps **…** and `…` spans in on/off and escapes all text. Unclosed
// markers stay literal.
func inline(s, on, off string, escape func(string) string) string {
	var b strings.Builder
	for s != "" {
		i := strings.IndexAny(s, "*`")
		if i < 0 {
			break
		}
		marker := "`"
		if s[i] == '*' {
			if !strings.HasPrefix(s[i:], "**") {
				b.WriteString(escape(s[:i+1]))
				s = s[i+1:]
				continue
			}
			marker = "**"
		}
		rest := s[i+len(marker):]
		end := strings.Index(rest, marker)
		if end < 0 {
			break
		}
		b.WriteString(escape(s[:i]) + on + escape(rest[:end]) + off)
		s = rest[end+len(marker):]
	}
	b.WriteString(escape(s))
	return b.String()
}
