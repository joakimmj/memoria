package cli

import (
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/joakimmj/memoria/docs"
)

func TestParsePage(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string // substring; "" means no error
		want    []pageSection
	}{
		{
			name: "blocks",
			src:  "# x(1)\n\n## NAME\n\nx - a\nthing\n\n## DESCRIPTION\n\n- one\n  more\n- two\n\n```\ncode  kept\n```\n",
			want: []pageSection{
				{heading: "NAME", blocks: []block{{kind: paragraphBlock, text: "x - a thing"}}},
				{heading: "DESCRIPTION", blocks: []block{
					{kind: listBlock, text: "one more"},
					{kind: listBlock, text: "two"},
					{kind: codeBlock, lines: []string{"code  kept"}},
				}},
			},
		},
		{name: "unknown section", src: "## NOTES\n", wantErr: "unknown section"},
		{name: "generated section", src: "## SYNOPSIS\n", wantErr: "generated"},
		{name: "text before sections", src: "hello\n", wantErr: "before the first section"},
		{name: "unclosed code", src: "## NAME\n```\nx\n", wantErr: "unclosed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parsePage("x.1.md", tt.src)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if p.name != "x" || p.section != "1" {
				t.Errorf("name, section = %q, %q; want x, 1", p.name, p.section)
			}
			if !slices.EqualFunc(p.sections, tt.want, func(a, b pageSection) bool {
				return a.heading == b.heading && slices.EqualFunc(a.blocks, b.blocks, func(x, y block) bool {
					return x.kind == y.kind && x.term == y.term && x.text == y.text && slices.Equal(x.lines, y.lines)
				})
			}) {
				t.Errorf("sections = %+v\nwant %+v", p.sections, tt.want)
			}
		})
	}
}

func TestInline(t *testing.T) {
	tests := []struct {
		in, plain, roff string
	}{
		{in: "plain", plain: "plain", roff: "plain"},
		{in: "a **b** c", plain: "a b c", roff: `a \fBb\fR c`},
		{in: "run `mia -h`", plain: "run mia -h", roff: `run \fBmia \-h\fR`},
		{in: "a * b", plain: "a * b", roff: "a * b"},
		{in: "open **bold", plain: "open **bold", roff: "open **bold"},
		{in: `back\slash`, plain: `back\slash`, roff: `back\eslash`},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := plainInline(tt.in); got != tt.plain {
				t.Errorf("plainInline = %q, want %q", got, tt.plain)
			}
			if got := roffInline(tt.in); got != tt.roff {
				t.Errorf("roffInline = %q, want %q", got, tt.roff)
			}
		})
	}
}

// TestEveryPageRenders: every embedded page parses and renders.
func TestEveryPageRenders(t *testing.T) {
	files, err := fs.Glob(docs.FS, "*.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			p, err := loadPage(f)
			if err != nil {
				t.Fatal(err)
			}
			if len(p.sections) == 0 || p.sections[0].heading != "NAME" {
				t.Errorf("first section is not NAME")
			}
			if renderText(p) == "" || renderRoff(p) == "" {
				t.Errorf("empty rendering")
			}
		})
	}
}
