// ABOUTME: Shares provider reset deadlines across projects without storing audit content.
// ABOUTME: Serializes atomic per-route JSON updates in a private runtime directory.
package harness

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// CooldownStore holds expiring route availability, never audit verdicts or budgets.
type CooldownStore struct{ Directory string }

type cooldownRecord struct {
	Version  int       `json:"version"`
	Harness  string    `json:"harness"`
	Provider string    `json:"provider,omitempty"`
	Model    string    `json:"model"`
	RetryAt  time.Time `json:"retry_at"`
}

// DefaultCooldownDirectory resolves runtime state separately from the installation.
// An explicit absolute override supports isolated runs without changing HOME.
func DefaultCooldownDirectory() (string, error) {
	if path := os.Getenv("SDLC_HARNESS_STATE_DIR"); path != "" {
		if !filepath.IsAbs(path) {
			return "", errors.New("SDLC_HARNESS_STATE_DIR must be absolute")
		}
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agent", "sdlc"), nil
}

func (s CooldownStore) path(agent AgentConfig) string {
	key := sha256.Sum256([]byte(agent.Harness + "\x00" + agent.Provider + "\x00" + agent.Model))
	return filepath.Join(s.Directory, "cooldowns", fmt.Sprintf("%x.json", key))
}

// Deadline returns the stored deadline, including an expired one. The caller
// decides eligibility against its clock; reading never mutates runtime state.
func (s CooldownStore) Deadline(agent AgentConfig) (time.Time, error) {
	file, err := os.Open(s.path(agent))
	if errors.Is(err, os.ErrNotExist) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("reading audit cooldown: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 4097))
	if err != nil {
		return time.Time{}, err
	}
	var record cooldownRecord
	if len(data) > 4096 || json.Unmarshal(data, &record) != nil || record.Version != 1 || record.RetryAt.IsZero() || record.Harness != agent.Harness || record.Provider != agent.Provider || record.Model != agent.Model {
		return time.Time{}, fmt.Errorf("invalid audit cooldown record %s", s.path(agent))
	}
	return record.RetryAt, nil
}

// Record retains the later deadline when concurrent invocations observe limits.
// Lock ownership ends on close or process exit; no provider runs under the lock.
func (s CooldownStore) Record(agent AgentConfig, retryAt time.Time) (returnErr error) {
	if retryAt.IsZero() {
		return errors.New("audit cooldown requires a reset deadline")
	}
	path := s.path(agent)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, lock.Close()) }()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	previous, err := s.Deadline(agent)
	if err != nil {
		return err
	}
	if !retryAt.After(previous) {
		return nil
	}
	data, err := json.Marshal(cooldownRecord{Version: 1, Harness: agent.Harness, Provider: agent.Provider, Model: agent.Model, RetryAt: retryAt.UTC()})
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".cooldown-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer func() {
		if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	_, writeErr := temporary.Write(data)
	if err := errors.Join(writeErr, temporary.Close()); err != nil {
		return err
	}
	return os.Rename(name, path)
}
