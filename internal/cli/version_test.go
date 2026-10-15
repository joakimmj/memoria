package cli

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want versionInfo
	}{
		{
			name: "no build info",
			info: nil,
			ok:   false,
			want: versionInfo{Version: "dev"},
		},
		{
			name: "go install at a tag",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}},
			ok:   true,
			want: versionInfo{Version: "v0.1.0"},
		},
		{
			name: "go build in a checkout",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "v0.1.1-0.20261015120000-abcdef123456+dirty"},
				Settings: []debug.BuildSetting{
					{Key: "vcs", Value: "git"},
					{Key: "vcs.revision", Value: "abcdef1234567890"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			ok:   true,
			want: versionInfo{Version: "v0.1.1-0.20261015120000-abcdef123456+dirty", Commit: "abcdef1234567890"},
		},
		{
			name: "devel build",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: versionInfo{Version: "dev"},
		},
		{
			name: "empty version",
			info: &debug.BuildInfo{},
			ok:   true,
			want: versionInfo{Version: "dev"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.info, tt.ok); got != tt.want {
				t.Errorf("resolveVersion() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
