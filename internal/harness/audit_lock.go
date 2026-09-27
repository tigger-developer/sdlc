// ABOUTME: Serializes short audit-record transactions across local processes.
// ABOUTME: Resolves aliases so reset and checkpoint writes use the same lock.
package harness

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func withAuditRecordLock(path string, update func(string) error) (returnErr error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if errors.Is(err, os.ErrNotExist) {
		// New records still share a canonical parent with every other writer.
		parent, parentErr := filepath.EvalSymlinks(filepath.Dir(path))
		if parentErr != nil {
			return parentErr
		}
		resolved, err = filepath.Join(parent, filepath.Base(path)), nil
	}
	if err != nil {
		return fmt.Errorf("resolving audit record: %w", err)
	}
	directory := filepath.Join(filepath.Dir(resolved), ".audit-locks")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	ignore, err := os.OpenFile(filepath.Join(directory, ".gitignore"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err == nil {
		_, writeErr := ignore.WriteString("*\n")
		if err := errors.Join(writeErr, ignore.Close()); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrExist) {
		return err
	}
	lockPath := filepath.Join(directory, filepath.Base(resolved)+".lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return fmt.Errorf("opening audit record lock: %w", err)
	}
	// Closing releases flock even after a failed transaction. Do not unlink the
	// lock file: another process may already be waiting on the same inode.
	defer func() { returnErr = errors.Join(returnErr, lock.Close()) }()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("locking audit record: %w", err)
	}
	return update(resolved)
}
