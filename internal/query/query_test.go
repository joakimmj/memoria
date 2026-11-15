package query

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/joakimmj/memoria/internal/atlas"
	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/config"
)

var (
	t1 = time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	t2 = t1.Add(time.Hour)
	t3 = t2.Add(time.Hour)
)

func source(name string, engrams map[string]time.Time, notDone, done []atlas.Task) Source {
	a := atlas.Atlas{Engrams: map[string]atlas.Entry{}, Backlinks: map[string][]string{}}
	for file, mod := range engrams {
		a.Engrams[file] = atlas.Entry{Title: file, Modified: mod}
	}
	a.Tasks = atlas.Tasks{NotDone: notDone, Done: done}
	return Source{Brain: brain.Brain{Name: name}, Atlas: a}
}

func TestEngrams(t *testing.T) {
	sources := []Source{
		source("work", map[string]time.Time{"old.md": t1, "new.md": t3, "same_b.md": t2}, nil, nil),
		source("home", map[string]time.Time{"same_a.md": t2, "same_z.md": t2}, nil, nil),
	}
	var got []string
	for _, s := range Engrams(sources) {
		got = append(got, s.Brain+"/"+s.File)
	}
	want := []string{"work/new.md", "home/same_a.md", "home/same_z.md", "work/same_b.md", "work/old.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Engrams order = %v, want %v", got, want)
	}
	if got := Engrams(nil); got == nil || len(got) != 0 {
		t.Errorf("Engrams(nil) = %#v, want an empty list", got)
	}
}

func TestTasks(t *testing.T) {
	work := source("work",
		map[string]time.Time{"a.md": t1, "b.md": t3, "c.md": t3},
		[]atlas.Task{{Engram: "a.md", Line: 3, Text: "a3"}, {Engram: "b.md", Line: 9, Text: "b9"}, {Engram: "b.md", Line: 2, Text: "b2"}, {Engram: "c.md", Line: 1, Text: "c1"}},
		[]atlas.Task{{Engram: "a.md", Line: 1, Text: "a1"}},
	)
	home := source("home",
		map[string]time.Time{"h.md": t2},
		[]atlas.Task{{Engram: "h.md", Line: 5, Text: "h5"}},
		nil,
	)
	tests := []struct {
		name      string
		filter    Filter
		want      []string // brain/text
		wantDone  int
		wantTotal int
	}{
		{
			name:      "every task: brain, engram newest first, then line",
			want:      []string{"home/h5", "work/b2", "work/b9", "work/c1", "work/a1", "work/a3"},
			wantDone:  1,
			wantTotal: 6,
		},
		{name: "state", filter: Filter{State: config.Done}, want: []string{"work/a1"}, wantDone: 1, wantTotal: 1},
		{name: "engram", filter: Filter{Engram: "a.md"}, want: []string{"work/a1", "work/a3"}, wantDone: 1, wantTotal: 2},
		{name: "both", filter: Filter{State: config.NotDone, Engram: "a.md"}, want: []string{"work/a3"}, wantTotal: 1},
		{name: "nothing matches", filter: Filter{Engram: "none.md"}, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tasks([]Source{work, home}, tt.filter)
			texts := []string{}
			for _, task := range got.Tasks {
				texts = append(texts, task.Brain+"/"+task.Text)
			}
			if got.Tasks == nil || !reflect.DeepEqual(texts, tt.want) || got.Done != tt.wantDone || got.Total != tt.wantTotal {
				t.Errorf("Tasks = %v (done %d, total %d), want %v (done %d, total %d)", texts, got.Done, got.Total, tt.want, tt.wantDone, tt.wantTotal)
			}
		})
	}
	got := Tasks([]Source{work}, Filter{Engram: "a.md", State: config.Done})
	want := Task{Brain: "work", Engram: "a.md", Line: 1, Text: "a1", State: config.Done}
	if len(got.Tasks) != 1 || got.Tasks[0] != want {
		t.Errorf("Tasks = %+v, want [%+v]", got.Tasks, want)
	}
}

func TestEngram(t *testing.T) {
	dir := t.TempDir()
	content := "<!-- SYNAPSES -->\n- **up:**\n***\n<!-- /SYNAPSES -->\n\n# A\ntext\n"
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ref := brain.Ref{Brain: brain.Brain{Name: "work", Location: dir}, File: "a.md"}
	a := atlas.Atlas{
		Engrams:   map[string]atlas.Entry{"a.md": {Title: "A"}, "b.md": {Title: "B"}},
		Backlinks: map[string][]string{"a.md": {"b.md"}},
	}
	cfg := config.Engrams{DateFormat: "DD.MM.YYYY", Filename: config.Filename{DateFormat: "YYYY-MM-DD"}}

	got, err := Engram(ref, a, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	want := Detail{
		Brain: "work", File: "a.md", Path: filepath.Join(dir, "a.md"),
		Entry: atlas.Entry{Title: "A"}, Backlinks: []string{"b.md"}, Content: content,
		DateFormat: "DD.MM.YYYY", FilenameDateFormat: "YYYY-MM-DD",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Engram =\n%+v\nwant\n%+v", got, want)
	}

	body, err := Engram(ref, a, cfg, true)
	if err != nil || body.Content != "# A\ntext\n" {
		t.Errorf("Engram body = %q, %v", body.Content, err)
	}

	if _, err := Engram(brain.Ref{Brain: ref.Brain, File: "c.md"}, a, cfg, false); !errors.Is(err, brain.ErrNoEngram) {
		t.Errorf("not in atlas: err = %v, want ErrNoEngram", err)
	}
	noBacklinks, _ := Engram(ref, atlas.Atlas{Engrams: a.Engrams}, cfg, false)
	if noBacklinks.Backlinks == nil {
		t.Errorf("Backlinks nil, want []")
	}
}
