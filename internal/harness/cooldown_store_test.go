// ABOUTME: Verifies persistent route isolation, expiry data and concurrent cooldown updates.
// ABOUTME: Uses only temporary files, without provider calls or home-directory state.
package harness

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestCooldownRuntimeIsIgnoredByGit(t *testing.T) {
	root := t.TempDir()
	init := exec.Command("git", "init", root)
	if out, err := init.CombinedOutput(); err != nil {
		t.Fatalf("init: %s %v", out, err)
	}
	store := CooldownStore{Directory: filepath.Join(root, ".sdlc")}
	agent := AgentConfig{Harness: "codex", Model: "fixture"}
	if err := store.Record(agent, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	// #nosec G204 -- fixed Git operation on a hashed fixture path under t.TempDir; no shell.
	check := exec.Command("git", "check-ignore", "--quiet", store.path(agent))
	check.Dir = root
	if out, err := check.CombinedOutput(); err != nil {
		t.Fatalf("runtime state not ignored: %s %v", out, err)
	}
}

func TestCooldownStoreRetainsLatestDeadlineAcrossInstances(t *testing.T) {
	store := CooldownStore{Directory: t.TempDir()}
	agent := AgentConfig{Harness: "claude", Model: "fixture"}
	base := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	var workers sync.WaitGroup
	for i := range 12 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := store.Record(agent, base.Add(time.Duration(i)*time.Minute)); err != nil {
				t.Error(err)
			}
		}()
	}
	workers.Wait()
	other := CooldownStore{Directory: store.Directory}
	got, err := other.Deadline(agent)
	if err != nil || !got.Equal(base.Add(11*time.Minute)) {
		t.Fatalf("deadline=%s err=%v", got, err)
	}
	for _, different := range []AgentConfig{{Harness: "claude", Model: "other"}, {Harness: "hermes", Model: "fixture"}, {Harness: "claude", Model: "fixture", Provider: "other"}} {
		got, err := other.Deadline(different)
		if err != nil || !got.IsZero() {
			t.Fatalf("unrelated route cooled: %v %s %v", different, got, err)
		}
	}
	info, err := os.Stat(store.path(agent))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("state permissions: %v %v", info, err)
	}
}

func TestCooldownStoreRejectsInvalidState(t *testing.T) {
	store := CooldownStore{Directory: t.TempDir()}
	agent := AgentConfig{Harness: "claude", Model: "fixture"}
	if err := store.Record(agent, time.Time{}); err == nil {
		t.Fatal("accepted missing deadline")
	}
	if err := os.MkdirAll(filepath.Dir(store.path(agent)), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, data := range []string{`{`, `{ "version": 2 }`, `{ "version": 1, "harness": "other", "retry_at": "2026-09-20T12:00:00Z" }`} {
		if err := os.WriteFile(store.path(agent), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Deadline(agent); err == nil {
			t.Fatal("accepted corrupt state")
		}
		if err := store.Record(agent, time.Now()); err == nil {
			t.Fatal("overwrote corrupt state")
		}
	}
}
