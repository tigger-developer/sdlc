package configenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesShellExpansionAndReturnsOnlyAllowedKeys(t *testing.T) {
	directory := t.TempDir()
	configuration := filepath.Join(directory, ".env")
	contents := "BASE=/srv/exodan\nexport SDLC_INFRA_CONTRACT=\"$BASE/PROJECT-INTEGRATION.md\"\nPRIVATE_TOKEN=secret\n"
	if err := os.WriteFile(configuration, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join("..", "..", "src", "libexec", "load-sdlc-env.sh")
	values, err := Load(wrapper, configuration, []string{"SDLC_INFRA_CONTRACT"})
	if err != nil {
		t.Fatal(err)
	}
	if values["SDLC_INFRA_CONTRACT"] != "/srv/exodan/PROJECT-INTEGRATION.md" {
		t.Fatalf("expanded contract = %q", values["SDLC_INFRA_CONTRACT"])
	}
	if _, exists := values["PRIVATE_TOKEN"]; exists {
		t.Fatal("unlisted secret was returned")
	}
}

func TestLoadDoesNotInheritAllowedKeysMissingFromFile(t *testing.T) {
	directory := t.TempDir()
	configuration := filepath.Join(directory, ".env")
	if err := os.WriteFile(configuration, []byte("UNRELATED=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SDLC_AUDIT_MODEL", "inherited-model")
	wrapper := filepath.Join("..", "..", "src", "libexec", "load-sdlc-env.sh")
	values, err := Load(wrapper, configuration, []string{"SDLC_AUDIT_MODEL"})
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := values["SDLC_AUDIT_MODEL"]; exists {
		t.Fatal("missing file value was inherited from the process environment")
	}
}
