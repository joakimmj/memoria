package atlas

import (
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/joakimmj/memoria/internal/config"
	"github.com/joakimmj/memoria/internal/markdown"
)

func defaults(t *testing.T) config.Config {
	t.Helper()
	c, err := config.Defaults()
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestIndex(t *testing.T) {
	const block = "<!-- SYNAPSES -->\n- **down:** [b](b.md)\n- **up:**\n- [ ] **see:** [c](c.md)\n***\n<!-- /SYNAPSES -->\n"
	tests := []struct {
		name, content string
		wantTitle     string
		wantSynapses  map[string][]string
		wantLinks     []string
		wantTasks     []stateTask
	}{
		{name: "empty", content: "", wantTitle: "x"},
		{name: "title is the first heading", content: "text\n## Sub\n# Top\n", wantTitle: "Sub"},
		{name: "empty heading skipped", content: "#\n# Real\n", wantTitle: "Real"},
		{
			name:         "block read, not body",
			content:      block + "\n# X\n- [ ] in body [d](d.md)\n",
			wantTitle:    "X",
			wantSynapses: map[string][]string{"down": {"b.md"}, "up": {}, "see": {"c.md"}},
			wantLinks:    []string{"d.md"},
			wantTasks:    []stateTask{{Task{"x.md", 9, "in body [d](d.md)"}, "not_done"}},
		},
		{
			name:         "frontmatter skipped, line numbers true",
			content:      "---\ntitle: # no\nl: [a](a.md)\n---\n" + block + "\n# X\n- [x] done\n- [X] also\n- [?] other\n- plain\n",
			wantTitle:    "X",
			wantSynapses: map[string][]string{"down": {"b.md"}, "up": {}, "see": {"c.md"}},
			wantTasks: []stateTask{
				{Task{"x.md", 13, "done"}, "done"},
				{Task{"x.md", 14, "also"}, "done"},
			},
		},
		{
			name:      "links: engrams only, once each, in order",
			content:   "[b](b.md) [a](./a.md) ![i](i.md) [w](https://x.org) [s](sub/s.md) [b again](b.md#top) `[c](c.md)`\n",
			wantTitle: "x",
			wantLinks: []string{"b.md", "a.md"},
		},
		{
			name:      "fenced code skipped",
			content:   "```sh\n# comment\n- [ ] not a task\n[a](a.md)\n```\n# Real\n- [ ] task\n",
			wantTitle: "Real",
			wantTasks: []stateTask{{Task{"x.md", 7, "task"}, "not_done"}},
		},
		{
			name:      "nested and ordered tasks",
			content:   "  - [ ] nested\n1. [x] ordered\n",
			wantTitle: "x",
			wantTasks: []stateTask{
				{Task{"x.md", 1, "nested"}, "not_done"},
				{Task{"x.md", 2, "ordered"}, "done"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, tasks := index("x.md", markdown.Lines(tt.content), defaults(t))
			wantSynapses := tt.wantSynapses
			if wantSynapses == nil {
				wantSynapses = map[string][]string{}
			}
			wantLinks := tt.wantLinks
			if wantLinks == nil {
				wantLinks = []string{}
			}
			if e.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", e.Title, tt.wantTitle)
			}
			if e.Synapses == nil || !maps.EqualFunc(e.Synapses, wantSynapses, slices.Equal) {
				t.Errorf("Synapses = %#v, want %q", e.Synapses, wantSynapses)
			}
			if e.Links == nil || !slices.Equal(e.Links, wantLinks) {
				t.Errorf("Links = %#v, want %q", e.Links, wantLinks)
			}
			if !reflect.DeepEqual(tasks, tt.wantTasks) && (len(tasks) != 0 || len(tt.wantTasks) != 0) {
				t.Errorf("tasks = %+v, want %+v", tasks, tt.wantTasks)
			}
		})
	}
}
