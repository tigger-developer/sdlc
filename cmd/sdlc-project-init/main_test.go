package main

import (
	"runtime/debug"
	"testing"
)

func TestSourceRevisionFromBuildInfo(t *testing.T) {
	clean := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef"},
			{Key: "vcs.modified", Value: "false"},
		},
	}
	if got := sourceRevisionFromBuildInfo(clean); got != "0123456789abcdef" {
		t.Fatalf("clean source revision = %q", got)
	}

	dirty := *clean
	dirty.Settings = append([]debug.BuildSetting(nil), clean.Settings...)
	dirty.Settings[1].Value = "true"
	if got := sourceRevisionFromBuildInfo(&dirty); got != "" {
		t.Fatalf("dirty source revision = %q, want unresolved", got)
	}

	release := &debug.BuildInfo{Main: debug.Module{Version: "v2.0.0"}}
	if got := sourceRevisionFromBuildInfo(release); got != "v2.0.0" {
		t.Fatalf("release source revision = %q", got)
	}
}

func TestSourceRevisionPrefersInjectedRelease(t *testing.T) {
	clean := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef"},
			{Key: "vcs.modified", Value: "false"},
		},
	}
	if got := sourceRevisionForBuildInfo(clean, "v2.0.1"); got != "v2.0.1" {
		t.Fatalf("source revision = %q, want injected release", got)
	}

	clean.Settings[1].Value = "true"
	if got := sourceRevisionForBuildInfo(clean, "v2.0.1"); got != "" {
		t.Fatalf("dirty source revision = %q, want unresolved", got)
	}
}
