package cli

import "runtime/debug"

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

func runVersion(Invocation) (any, error) {
	info, ok := debug.ReadBuildInfo()
	return resolveVersion(info, ok), nil
}

func resolveVersion(info *debug.BuildInfo, ok bool) versionInfo {
	if !ok {
		return versionInfo{Version: "dev"}
	}
	return versionInfo{
		Version: moduleVersion(info.Main.Version),
		Commit:  setting(info.Settings, "vcs.revision"),
	}
}

func moduleVersion(v string) string {
	if v == "" || v == "(devel)" {
		return "dev"
	}
	return v
}

func setting(settings []debug.BuildSetting, key string) string {
	for _, s := range settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}
