package installer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Trash exact managed paths, not symlink targets. Configuration merges retain
// their separate adjacent-backup and restoration path.
func trashArtifact(output io.Writer, path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "trash", absolute)
	command.WaitDelay = time.Second
	diagnostic, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("trash %q: %w: %s", absolute, err, strings.TrimSpace(string(diagnostic)))
	}
	if _, err := os.Lstat(absolute); !os.IsNotExist(err) {
		return fmt.Errorf("trash did not remove the original path %q (inspection: %v)", absolute, err)
	}
	fmt.Fprintf(output, "Trashed: %s (recoverable through Trash)\n", absolute)
	return nil
}
