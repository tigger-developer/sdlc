// ABOUTME: Installs canonical SDLC assets without per-user adapters or discovery.
// ABOUTME: Leaves all home links and harness configuration to declarative management.
package installer

import (
	"errors"
	"fmt"
	"os"
)

func RunBare(options Options) error {
	options = withDefaultIO(options)
	root := options.InstallRoot
	if root == "" {
		root = canonicalInstallRoot
	}
	source, err := validateSource(options.Source)
	if err != nil {
		return err
	}
	plan, err := planSharedInstallationMode(source, "", root, options.Release, true)
	if err != nil {
		return err
	}
	printInstallationPlan(options.Output, plan, options.Apply)
	if !options.Apply {
		return nil
	}
	if err := ensureBareRoot(root); err != nil {
		return err
	}
	if err := applyInstallation(plan, options.Output); err != nil {
		return err
	}
	if err := ensureSharedDirectoryModes(root); err != nil {
		return err
	}
	verified, err := planSharedInstallationMode(source, "", root, options.Release, true)
	if err != nil {
		return err
	}
	if installationHasChanges(verified) {
		return errors.New("bare installation verification still reports changes")
	}
	if err := verifyCanonicalMain(root); err != nil {
		return err
	}
	_, err = fmt.Fprintln(options.Output, "Canonical SDLC assets installed; no home links or harness configuration changed.")
	return err
}

func ensureBareRoot(root string) error {
	info, err := os.Lstat(root)
	if err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("SDLC installation root %q is not a regular directory", root)
		}
		return checkInstallRootWritable(root)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("cannot create %s without privileges; pre-create it with ownership for the installing user before --bare: %w", root, err)
	}
	return checkInstallRootWritable(root)
}
