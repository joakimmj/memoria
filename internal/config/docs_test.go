package config

import (
	"io/fs"
	"reflect"
	"strings"
	"testing"

	"github.com/joakimmj/memoria/docs"
)

func TestEveryKeyDocumented(t *testing.T) {
	page, err := fs.ReadFile(docs.FS, "mia-config.5.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys(reflect.TypeFor[Config](), "") {
		if !strings.Contains(string(page), "`"+key+"`") {
			t.Errorf("mia-config(5) does not document `%s`", key)
		}
	}
}

// keys lists typ's dotted JSON keys; map entries are <name>.
func keys(typ reflect.Type, prefix string) []string {
	var out []string
	for f := range typ.Fields() {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		key := prefix + name
		switch ft := f.Type; ft.Kind() {
		case reflect.Struct:
			out = append(out, keys(ft, key+".")...)
		case reflect.Map:
			out = append(out, key+".<name>")
			out = append(out, keys(ft.Elem(), key+".<name>.")...)
		default:
			out = append(out, key)
		}
	}
	return out
}

// Decoding never fails on a missing key, so check defaults.json here.
func TestDefaultsComplete(t *testing.T) {
	for _, key := range keys(reflect.TypeFor[Config](), "") {
		key, _, _ = strings.Cut(key, ".<name>")
		if !has(defaultTree, strings.Split(key, ".")) {
			t.Errorf("defaults.json has no `%s`", key)
		}
	}
}

func has(tree any, path []string) bool {
	for _, k := range path {
		m, _ := tree.(map[string]any)
		v, ok := m[k]
		if !ok {
			return false
		}
		tree = v
	}
	return true
}
