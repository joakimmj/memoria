package atlas

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/joakimmj/memoria/internal/brain"
)

func newBrain(t *testing.T, files map[string]string) brain.Brain {
	t.Helper()
	b := brain.Brain{Name: "work", Location: t.TempDir()}
	for f, content := range files {
		writeFile(t, b, f, content)
	}
	return b
}

func writeFile(t *testing.T, b brain.Brain, file, content string) {
	t.Helper()
	path := filepath.Join(b.Location, file)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func stored(t *testing.T, b brain.Brain) Atlas {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(b.Location, File))
	if err != nil {
		t.Fatal(err)
	}
	var a Atlas
	if err := json.Unmarshal(data, &a); err != nil {
		t.Fatal(err)
	}
	return a
}

// backdate makes a write by Refresh visible as a newer mtime.
func backdate(t *testing.T, b brain.Brain) time.Time {
	t.Helper()
	past := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(filepath.Join(b.Location, File), past, past); err != nil {
		t.Fatal(err)
	}
	return past
}

func atlasModTime(t *testing.T, b brain.Brain) time.Time {
	t.Helper()
	info, err := os.Stat(filepath.Join(b.Location, File))
	if err != nil {
		t.Fatal(err)
	}
	return info.ModTime()
}

func TestRefresh(t *testing.T) {
	cfg := defaults(t)
	b := newBrain(t, map[string]string{
		"a.md":       "<!-- SYNAPSES -->\n- **down:** [b](b.md)\n***\n<!-- /SYNAPSES -->\n\n# A\n- [ ] one\n[gone](gone.md)\n",
		"b.md":       "# B\n- [x] two\n- [ ] three\n",
		"notes.txt":  "not an engram",
		"sub/c.md":   "# not top-level\n",
		".hidden.md": "# hidden engram\n",
	})
	if err := os.Symlink(filepath.Join(b.Location, "b.md"), filepath.Join(b.Location, "link.md")); err != nil {
		t.Fatal(err)
	}

	got, err := Refresh(b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, stored(t, b)) {
		t.Errorf("Refresh answer and stored atlas differ:\n%+v\n%+v", got, stored(t, b))
	}
	if files := slices.Sorted(maps.Keys(got.Engrams)); !slices.Equal(files, []string{".hidden.md", "a.md", "b.md"}) {
		t.Errorf("engrams = %v", files)
	}
	if e := got.Engrams["a.md"]; e.Title != "A" || e.Hash == "" || e.Modified.IsZero() || e.Modified.Location() != time.UTC {
		t.Errorf("a.md = %+v", e)
	}
	wantBack := map[string][]string{"b.md": {"a.md"}, "gone.md": {"a.md"}}
	if !reflect.DeepEqual(got.Backlinks, wantBack) {
		t.Errorf("backlinks = %v, want %v", got.Backlinks, wantBack)
	}
	wantTasks := Tasks{
		NotDone: []Task{{"a.md", 7, "one"}, {"b.md", 3, "three"}},
		Done:    []Task{{"b.md", 2, "two"}},
	}
	if !reflect.DeepEqual(got.Tasks, wantTasks) {
		t.Errorf("tasks = %+v, want %+v", got.Tasks, wantTasks)
	}
	if info, _ := os.Stat(filepath.Join(b.Location, File)); info.Mode().Perm() != 0o644 {
		t.Errorf("atlas mode = %v, want 0644", info.Mode().Perm())
	}

	t.Run("nothing changed, nothing written", func(t *testing.T) {
		past := backdate(t, b)
		if _, err := Refresh(b, cfg); err != nil {
			t.Fatal(err)
		}
		if !atlasModTime(t, b).Equal(past) {
			t.Errorf("atlas rewritten with nothing changed")
		}
	})

	t.Run("changed engram re-parsed, others kept", func(t *testing.T) {
		writeFile(t, b, "b.md", "# B2\n- [x] two\n")
		got, err := Refresh(b, cfg)
		if err != nil {
			t.Fatal(err)
		}
		want := Tasks{NotDone: []Task{{"a.md", 7, "one"}}, Done: []Task{{"b.md", 2, "two"}}}
		if got.Engrams["b.md"].Title != "B2" || !reflect.DeepEqual(got.Tasks, want) {
			t.Errorf("b.md = %+v, tasks = %+v", got.Engrams["b.md"], got.Tasks)
		}
	})

	t.Run("vanished engram dropped with its tasks and links", func(t *testing.T) {
		if err := os.Remove(filepath.Join(b.Location, "a.md")); err != nil {
			t.Fatal(err)
		}
		got, err := Refresh(b, cfg)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.Engrams["a.md"]; ok || len(got.Backlinks) != 0 || len(got.Tasks.NotDone) != 0 {
			t.Errorf("a.md still indexed: %+v", got)
		}
	})

	t.Run("config change re-parses every engram", func(t *testing.T) {
		changed := cfg
		changed.Engrams.TaskMarkers.Done = []string{"[x]", "[X]", "[-]"}
		writeFile(t, b, "d.md", "- [-] cancelled\n")
		if _, err := Refresh(b, cfg); err != nil {
			t.Fatal(err)
		}
		got, err := Refresh(b, changed)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(got.Tasks.Done, Task{"d.md", 1, "cancelled"}) {
			t.Errorf("tasks = %+v, want d.md's [-] done", got.Tasks)
		}
	})

	t.Run("corrupt atlas rebuilt", func(t *testing.T) {
		writeFile(t, b, File, "{not json")
		got, err := Refresh(b, cfg)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Engrams) == 0 || !reflect.DeepEqual(got, stored(t, b)) {
			t.Errorf("not rebuilt: %+v", got)
		}
	})
}

func TestRefreshEmptyBrain(t *testing.T) {
	b := newBrain(t, nil)
	if _, err := Refresh(b, defaults(t)); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(b.Location, File))
	if err != nil {
		t.Fatal(err)
	}
	var a map[string]any
	if err := json.Unmarshal(data, &a); err != nil {
		t.Fatal(err)
	}
	if a["engrams"] == nil || a["backlinks"] == nil || a["tasks"].(map[string]any)["done"] == nil {
		t.Errorf("empty atlas has nulls: %s", data)
	}
}

func TestRefreshMissingFolder(t *testing.T) {
	b := brain.Brain{Name: "gone", Location: filepath.Join(t.TempDir(), "gone")}
	if _, err := Refresh(b, defaults(t)); err == nil {
		t.Error("err = nil, want one for a missing folder")
	}
}
