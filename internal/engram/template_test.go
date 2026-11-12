package engram

import (
	"errors"
	"testing"
)

func TestRender(t *testing.T) {
	v := vars{title: "Project X", date: "2026-08-01"}
	tests := []struct {
		name     string
		template string
		v        vars
		body     string
		want     string
		wantPos  Position
		wantErr  error
	}{
		{name: "default template", template: "# %title%\n\n%body%", v: v, want: "# Project X\n\n", wantPos: Position{3, 0}},
		{name: "body inserted", template: "# %title%\n\n%body%", v: v, body: "one\ntwo", want: "# Project X\n\none\ntwo", wantPos: Position{3, 0}},
		{name: "body mid-line", template: "a %body% b", v: v, body: "X", want: "a X b", wantPos: Position{1, 2}},
		{name: "date before body counts expanded", template: "%date% %body%", v: v, want: "2026-08-01 ", wantPos: Position{1, 11}},
		{name: "col counts bytes", template: "# %title% %body%", v: vars{title: "Ærlig"}, want: "# Ærlig ", wantPos: Position{1, 9}},
		{name: "no body placeholder appends", template: "# %title%\n", v: v, body: "hi", want: "# Project X\nhi", wantPos: Position{2, 0}},
		{name: "no body placeholder, no body", template: "%title%", v: v, want: "Project X", wantPos: Position{1, 9}},
		{name: "placeholders everywhere", template: "%title% %title% %date%", v: v, want: "Project X Project X 2026-08-01", wantPos: Position{1, 30}},
		{name: "title is not expanded again", template: "%title%%body%", v: vars{title: "%date%", date: "D"}, want: "%date%", wantPos: Position{1, 6}},
		{name: "empty title", template: "# %title%\n\n%body%", v: vars{}, want: "# \n\n", wantPos: Position{3, 0}},
		{name: "two body placeholders", template: "%body% %body%", v: v, wantErr: errManyBodies},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, pos, err := render(tt.template, tt.v, tt.body)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want || pos != tt.wantPos {
				t.Errorf("render = %q %+v, want %q %+v", got, pos, tt.want, tt.wantPos)
			}
		})
	}
}
