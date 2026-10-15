package cli

import "runtime/debug"

// versionInfo is the result of `mia version`.
type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

func runVersion(Invocation) (any, error) {
	info, ok := debug.ReadBuildInfo()
	return resolveVersion(info, ok), nil
}

// resolveVersion reads the version and commit from the build info.
func resolveVersion(info *debug.BuildInfo, ok bool) versionInfo {
	if !ok {
		return versionInfo{Version: "dev"}
	}
	return versionInfo{
		Version: moduleVersion(info.Main.Version),
		Commit:  setting(info.Settings, "vcs.revision"),
	}
}

// moduleVersion is v, or "dev" when the build has no real version.
func moduleVersion(v string) string {
	if v == "" || v == "(devel)" {
		return "dev"
	}
	return v
}

// setting returns the value of the build setting key, or "".
func setting(settings []debug.BuildSetting, key string) string {
	for _, s := range settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}
