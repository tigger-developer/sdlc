// ABOUTME: Builds the complete canonical SDLC tree from the deployment mapping.
// ABOUTME: Preserves validated operator configuration while replacing stale framework assets.
package installer

import (
	"github.com/tigger-developer/sdlc/internal/deployment"
	"os"
	"path/filepath"
)

func withinRoot(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && filepath.IsLocal(relative)
}
func (plan *installationPlan) prepareCanonical(root string) error {
	plan.root = root
	plan.tree = deployment.Tree{}
	for _, item := range plan.syncs {
		if !withinRoot(item.destination, root) {
			continue
		}
		relative, err := filepath.Rel(root, item.destination)
		if err != nil {
			return err
		}
		if err := plan.tree.Copy(item.source, relative); err != nil {
			return err
		}
	}
	// sdlc.yaml belongs to the operator. A proposed schema/release migration takes
	// precedence, otherwise preserve its exact bytes and permissions.
	config := filepath.Join(root, "sdlc.yaml")
	data, mode, exists, err := readOptionalRegularFile(config)
	if err != nil {
		return err
	}
	for _, change := range plan.managedConfigurations {
		if change.path == config {
			data = change.contents
			mode = change.mode
			exists = true
		}
	}
	if exists {
		if _, _, err := globalConfigurationDocument(config, data); err != nil {
			return err
		}
		if err := plan.tree.Add("sdlc.yaml", "preserved operator configuration", data, mode); err != nil {
			return err
		}
	}
	if err := plan.applySkillPreferences(data); err != nil {
		return err
	}
	files, err := goosePluginContents()
	if err != nil {
		return err
	}
	for name, data := range files {
		if err := plan.tree.Add(filepath.Join("plugins", "sdlc-command-guard-goose", name), "generated Goose plugin", data, 0o644); err != nil {
			return err
		}
	}
	same, err := plan.tree.Matches(root)
	if err != nil {
		return err
	}
	plan.treeChanged = !same
	return nil
}

func prepareRecovery(root string, apply bool) error {
	pending, err := deployment.Pending(root)
	if err != nil {
		return err
	}
	if !pending {
		return nil
	}
	if !apply {
		return &os.PathError{Op: "installation recovery required; rerun without preview", Path: root, Err: os.ErrInvalid}
	}
	return deployment.Recover(root)
}
