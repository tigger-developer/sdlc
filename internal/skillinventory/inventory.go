// ABOUTME: Records installer-owned skills and plans recoverable retirement.
// ABOUTME: Rejects changed retired assets and never adopts destination-only skills.
package skillinventory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type record struct {
	Version int               `json:"version"`
	Owner   string            `json:"owner"`
	Skills  map[string]string `json:"skills"`
}

type Retirement struct{ Path, Fingerprint string }
type Plan struct {
	Path        string
	Contents    []byte
	Retirements []Retirement
	Changed     bool
}

// Build uses source fingerprints for copies and exact canonical targets for links.
// Each discovery directory owns a separate inventory, so unselected homes are untouched.
func Build(source, destination, canonical string, names []string, owner string) (Plan, error) {
	p := Plan{Path: filepath.Join(destination, "."+owner+"-inventory.json")}
	if info, err := os.Lstat(destination); err == nil {
		if !info.IsDir() {
			return p, fmt.Errorf("skill inventory root %q must be a real directory", destination)
		}
	} else if !os.IsNotExist(err) {
		return p, err
	}
	current := record{Version: 1, Owner: owner, Skills: map[string]string{}}
	for _, name := range names {
		if !validName(name) {
			return p, fmt.Errorf("invalid skill name %q", name)
		}
		fingerprint := "link:" + filepath.Join(canonical, name)
		if canonical == "" {
			var err error
			fingerprint, err = fingerprintPath(filepath.Join(source, name))
			if err != nil {
				return p, err
			}
			if !strings.HasPrefix(fingerprint, "tree:") {
				return p, fmt.Errorf("source skill %q must be a directory", name)
			}
		}
		current.Skills[name] = fingerprint
	}
	old := record{}
	if info, statErr := os.Lstat(p.Path); statErr == nil && !info.Mode().IsRegular() {
		return p, fmt.Errorf("inventory %q must be a regular file", p.Path)
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return p, statErr
	}
	previous, err := os.ReadFile(p.Path)
	if err != nil && !os.IsNotExist(err) {
		return p, err
	}
	if err == nil {
		if err := json.Unmarshal(previous, &old); err != nil {
			return p, fmt.Errorf("read skill inventory %q: %w", p.Path, err)
		}
		if old.Version != 1 || old.Owner != owner || old.Skills == nil {
			return p, fmt.Errorf("unsupported skill inventory %q", p.Path)
		}
	}
	for name, fingerprint := range old.Skills {
		if !validName(name) || !validFingerprint(fingerprint, canonical, name) {
			return p, fmt.Errorf("invalid inventory entry %q in %s", name, p.Path)
		}
		if _, exists := current.Skills[name]; exists {
			continue
		}
		retirement := Retirement{filepath.Join(destination, name), fingerprint}
		if err := retirement.Validate(); err != nil {
			return p, err
		}
		if _, err := os.Lstat(retirement.Path); err == nil {
			p.Retirements = append(p.Retirements, retirement)
		} else if !os.IsNotExist(err) {
			return p, err
		}
	}
	sort.Slice(p.Retirements, func(i, j int) bool { return p.Retirements[i].Path < p.Retirements[j].Path })
	p.Contents, err = json.MarshalIndent(current, "", "  ")
	if err != nil {
		return p, err
	}
	p.Contents = append(p.Contents, '\n')
	p.Changed = !bytes.Equal(previous, p.Contents)
	return p, nil
}

func validName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\") && !strings.HasPrefix(name, ".")
}
func validFingerprint(value, canonical, name string) bool {
	if canonical != "" {
		return value == "link:"+filepath.Join(canonical, name)
	}
	if !strings.HasPrefix(value, "tree:") {
		return false
	}
	digest, err := hex.DecodeString(strings.TrimPrefix(value, "tree:"))
	return err == nil && len(digest) == sha256.Size
}

func (r Retirement) Validate() error {
	if _, err := os.Lstat(r.Path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	actual, err := fingerprintPath(r.Path)
	if err != nil {
		return err
	}
	if actual != r.Fingerprint {
		return fmt.Errorf("retired skill %q differs from its installed inventory; preserve or review it before retrying", r.Path)
	}
	return nil
}

func fingerprintPath(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		return "link:" + target, err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("skill %q is not a directory or link", path)
	}
	files := map[string]string{}
	err = filepath.WalkDir(path, func(p string, e fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if e.IsDir() {
			return nil
		}
		name := e.Name()
		if name == ".DS_Store" || name == "Thumbs.db" || name == "desktop.ini" || strings.HasPrefix(name, "._") {
			return nil
		}
		if !e.Type().IsRegular() {
			return fmt.Errorf("nonregular skill entry %q cannot establish ownership", p)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(data)
		files[filepath.ToSlash(relative)] = hex.EncodeToString(digest[:])
		return nil
	})
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(files)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return "tree:" + hex.EncodeToString(digest[:]), nil
}
