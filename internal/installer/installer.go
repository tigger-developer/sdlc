package installer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	agentAuto    = "auto"
	agentClaude  = "claude"
	agentCodex   = "codex"
	agentCopilot = "copilot"
	agentHermes  = "hermes"
	agentCustom  = "custom"
)

type providerDefinition struct {
	name        string
	commandPath string
	skills      bool
}

var providerDefinitions = []providerDefinition{
	{name: agentClaude, commandPath: "commands", skills: true},
	{name: agentCodex, commandPath: "prompts-commands"},
	{name: agentCopilot, commandPath: "prompts-commands", skills: true},
	{name: agentHermes, skills: true},
}

var claudeDeniedCommands = []string{
	"Bash(rm:*)",
	"Bash(sed:*)",
	"Bash(awk:*)",
	"Bash(source:*)",
	"Bash(python:*)",
	"Bash(python3:*)",
}

const (
	codexPythonRulesStart = "# BEGIN SDLC MANAGED PYTHON RULES"
	codexPythonRulesEnd   = "# END SDLC MANAGED PYTHON RULES"
)

type Options struct {
	Agent     string
	AgentHome string
	Source    string
	Apply     bool
	Configure bool
	Input     io.Reader
	Output    io.Writer
}

type configurationChange struct {
	path        string
	beforeLabel string
	afterLabel  string
	contents    []byte
	mode        os.FileMode
}

type managedSync struct {
	source            string
	destination       string
	excludeGit        bool
	needsSync         bool
	backupExisting    bool
	destinationExists bool
}

type installationPlan struct {
	syncs []managedSync
}

func Run(options Options) error {
	options = withDefaultIO(options)
	source, err := validateSource(options.Source)
	if err != nil {
		return err
	}
	agent, agentHome, err := resolveTarget(options.Agent, options.AgentHome, source)
	if err != nil {
		return err
	}
	plan, err := planInstallation(agent, source, agentHome)
	if err != nil {
		return err
	}

	printHeading(options.Output, agent, source, agentHome, options.Apply, options.Configure)
	printInstallationPlan(options.Output, plan, options.Apply)
	change, err := analyseConfiguration(agent, agentHome, source, options.Output)
	if err != nil {
		return err
	}
	if options.Apply {
		if err := applyInstallation(plan, options.Output); err != nil {
			return err
		}
		verified, verifyErr := planInstallation(agent, source, agentHome)
		if verifyErr != nil {
			return fmt.Errorf("verifying installation: %w", verifyErr)
		}
		if installationHasChanges(verified) {
			return errors.New("installation verification still reports changes")
		}
		fmt.Fprintln(options.Output, "Installation: synchronized all planned copies.")
	}
	if options.Configure {
		return offerConfigurationChange(change, options.Input, options.Output)
	}
	if change != nil {
		fmt.Fprintln(options.Output, "Configuration: recommendation only; re-run with --configure to review and confirm it.")
	}
	return nil
}

func RunInteractive(sourcePath, userHome string, input io.Reader, output io.Writer) error {
	if input == nil {
		input = os.Stdin
	}
	if output == nil {
		output = os.Stdout
	}
	source, err := validateSource(sourcePath)
	if err != nil {
		return err
	}
	if userHome == "" {
		userHome, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolving user home: %w", err)
		}
	}
	userHome, err = filepath.Abs(userHome)
	if err != nil {
		return fmt.Errorf("resolving user home %q: %w", userHome, err)
	}
	agents, err := detectedAgents(userHome)
	if err != nil {
		return err
	}
	plan := installationPlan{}
	commonHome := filepath.Join(userHome, ".agents")
	if directoryExists(commonHome) {
		shared, planErr := planSharedInstallation(source, commonHome)
		if planErr != nil {
			return planErr
		}
		plan.syncs = append(plan.syncs, shared.syncs...)
	}
	for _, agent := range agents {
		agentHome := filepath.Join(userHome, "."+agent)
		provider, planErr := planProviderInstallation(agent, source, agentHome)
		if planErr != nil {
			return planErr
		}
		plan.syncs = append(plan.syncs, provider.syncs...)
	}

	fmt.Fprintln(output, "SDLC installer: INTERACTIVE")
	if len(agents) == 0 {
		fmt.Fprintln(output, "Detected agents: none")
	} else {
		fmt.Fprintf(output, "Detected agents: %s\n", strings.Join(agents, ", "))
	}
	if !installationHasChanges(plan) {
		fmt.Fprintln(output, "All detected SDLC copies are current.")
		return nil
	}
	printInstallationPlan(output, plan, false)
	accepted, confirmErr := confirm(bufio.NewReader(input), output, "Deploy all listed SDLC changes? [yes/no]: ")
	if confirmErr != nil {
		return confirmErr
	}
	if !accepted {
		fmt.Fprintln(output, "Deployment declined; no changes were made.")
		return nil
	}
	if err := applyInstallation(plan, output); err != nil {
		return err
	}
	verified, verifyErr := planDetectedInstallation(source, userHome, agents)
	if verifyErr != nil {
		return verifyErr
	}
	if installationHasChanges(verified) {
		return errors.New("deployment verification still reports changes")
	}
	fmt.Fprintln(output, "All listed SDLC changes were installed.")
	return nil
}

func planDetectedInstallation(source, userHome string, agents []string) (installationPlan, error) {
	plan := installationPlan{}
	commonHome := filepath.Join(userHome, ".agents")
	if directoryExists(commonHome) {
		shared, err := planSharedInstallation(source, commonHome)
		if err != nil {
			return installationPlan{}, err
		}
		plan.syncs = append(plan.syncs, shared.syncs...)
	}
	for _, agent := range agents {
		provider, err := planProviderInstallation(agent, source, filepath.Join(userHome, "."+agent))
		if err != nil {
			return installationPlan{}, err
		}
		plan.syncs = append(plan.syncs, provider.syncs...)
	}
	return plan, nil
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func confirm(reader *bufio.Reader, output io.Writer, prompt string) (bool, error) {
	fmt.Fprint(output, prompt)
	response, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("reading installation confirmation: %w", err)
	}
	return strings.EqualFold(strings.TrimSpace(response), "yes"), nil
}

func detectedAgents(userHome string) ([]string, error) {
	var detected []string
	for _, provider := range providerDefinitions {
		path := filepath.Join(userHome, "."+provider.name)
		info, err := os.Stat(path)
		switch {
		case errors.Is(err, os.ErrNotExist):
			continue
		case err != nil:
			return nil, fmt.Errorf("detecting agent home %q: %w", path, err)
		case info.IsDir():
			detected = append(detected, provider.name)
		}
	}
	return detected, nil
}

func installationHasChanges(plan installationPlan) bool {
	for _, sync := range plan.syncs {
		if sync.needsSync {
			return true
		}
	}
	return false
}

func withDefaultIO(options Options) Options {
	if options.Input == nil {
		options.Input = os.Stdin
	}
	if options.Output == nil {
		options.Output = os.Stdout
	}
	if options.Agent == "" {
		options.Agent = agentAuto
	}
	return options
}

func validateSource(source string) (string, error) {
	if source == "" {
		return "", errors.New("source path is required")
	}
	absolute, err := filepath.Abs(source)
	if err != nil {
		return "", fmt.Errorf("resolving source %q: %w", source, err)
	}
	for _, required := range []string{"MAIN.md", "README.md"} {
		path := filepath.Join(absolute, required)
		info, statErr := os.Stat(path)
		if statErr != nil {
			return "", fmt.Errorf("validating source %q: %w", path, statErr)
		}
		if info.IsDir() {
			return "", fmt.Errorf("validating source %q: expected a file", path)
		}
	}
	return filepath.Clean(absolute), nil
}

func resolveTarget(agent, agentHome, source string) (string, string, error) {
	requested := strings.ToLower(strings.TrimSpace(agent))
	if requested == "" {
		requested = agentAuto
	}
	if agentHome != "" {
		resolvedHome, err := filepath.Abs(agentHome)
		if err != nil {
			return "", "", fmt.Errorf("resolving agent home %q: %w", agentHome, err)
		}
		agentHome = filepath.Clean(resolvedHome)
	}
	if requested == agentAuto {
		return inferTarget(agentHome, source)
	}
	if !knownAgent(requested) {
		return "", "", fmt.Errorf("unsupported agent %q; use auto, claude, codex, copilot, hermes, or custom", agent)
	}
	if agentHome == "" {
		if requested == agentCustom {
			return "", "", errors.New("--agent custom requires --agent-home")
		}
		defaultHome, err := defaultAgentHome(requested)
		if err != nil {
			return "", "", err
		}
		agentHome = defaultHome
	}
	return requested, agentHome, nil
}

func inferTarget(agentHome, source string) (string, string, error) {
	if agentHome != "" {
		if agent := agentFromHome(agentHome); agent != "" {
			return agent, agentHome, nil
		}
		return "", "", fmt.Errorf("cannot infer agent from %q; supply --agent", agentHome)
	}
	if filepath.Base(source) == "sdlc" {
		parent := filepath.Dir(source)
		if agent := agentFromHome(parent); agent != "" {
			return agent, parent, nil
		}
	}
	return "", "", errors.New("cannot infer agent; supply --agent or --agent-home")
}

func knownAgent(agent string) bool {
	if agent == agentCustom {
		return true
	}
	_, found := providerDefinitionFor(agent)
	return found
}

func providerDefinitionFor(agent string) (providerDefinition, bool) {
	for _, provider := range providerDefinitions {
		if provider.name == agent {
			return provider, true
		}
	}
	return providerDefinition{}, false
}

func agentFromHome(path string) string {
	switch strings.TrimPrefix(strings.ToLower(filepath.Base(path)), ".") {
	case agentClaude:
		return agentClaude
	case agentCodex:
		return agentCodex
	case agentCopilot:
		return agentCopilot
	case agentHermes:
		return agentHermes
	default:
		return ""
	}
}

func defaultAgentHome(agent string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving user home: %w", err)
	}
	return filepath.Join(home, "."+agent), nil
}

func planInstallation(agent, source, agentHome string) (installationPlan, error) {
	commonHome := filepath.Join(filepath.Dir(agentHome), ".agents")
	if filepath.Clean(source) == filepath.Clean(filepath.Join(agentHome, "sdlc")) {
		return installationPlan{}, fmt.Errorf("source %q is inside the provider home; move the staging clone outside the provider home before installing the provider copy", source)
	}
	plan := installationPlan{}
	if directoryExists(commonHome) {
		shared, err := planSharedInstallation(source, commonHome)
		if err != nil {
			return installationPlan{}, err
		}
		plan.syncs = append(plan.syncs, shared.syncs...)
	}
	provider, err := planProviderInstallation(agent, source, agentHome)
	if err != nil {
		return installationPlan{}, err
	}
	plan.syncs = append(plan.syncs, provider.syncs...)
	return plan, nil
}

func planSharedInstallation(source, commonHome string) (installationPlan, error) {
	liveSDLC := filepath.Join(commonHome, "sdlc")
	if sameLexicalPath(source, liveSDLC) {
		return installationPlan{}, fmt.Errorf("source %q is the live SDLC directory; use a separate staging clone", source)
	}
	sync, err := planDirectorySync(source, liveSDLC, true)
	if err != nil {
		return installationPlan{}, err
	}
	skills, err := planChildDirectories(filepath.Join(source, "skills"), filepath.Join(commonHome, "skills"))
	if err != nil {
		return installationPlan{}, err
	}
	return installationPlan{syncs: append([]managedSync{sync}, skills...)}, nil
}

func planProviderInstallation(agent, source, agentHome string) (installationPlan, error) {
	if filepath.Clean(source) == filepath.Clean(filepath.Join(agentHome, "sdlc")) {
		return installationPlan{}, fmt.Errorf("source %q is inside the provider home; move the staging clone outside the provider home before installing the provider copy", source)
	}
	base, err := planDirectorySync(source, filepath.Join(agentHome, "sdlc"), true)
	if err != nil {
		return installationPlan{}, err
	}
	plan := installationPlan{syncs: []managedSync{base}}
	if agent == agentCustom {
		return plan, nil
	}
	provider, found := providerDefinitionFor(agent)
	if !found {
		return installationPlan{}, fmt.Errorf("planning unsupported agent %q", agent)
	}
	if provider.commandPath != "" {
		commands, planErr := planDirectorySync(filepath.Join(source, "commands"), filepath.Join(agentHome, provider.commandPath), false)
		if planErr != nil {
			return installationPlan{}, planErr
		}
		plan.syncs = append(plan.syncs, commands)
	}
	if provider.skills {
		skills, planErr := planChildDirectories(filepath.Join(source, "skills"), filepath.Join(agentHome, "skills"))
		if planErr != nil {
			return installationPlan{}, planErr
		}
		plan.syncs = append(plan.syncs, skills...)
	}
	return plan, nil
}

func planChildDirectories(sourceDirectory, destinationDirectory string) ([]managedSync, error) {
	entries, err := os.ReadDir(sourceDirectory)
	if err != nil {
		return nil, fmt.Errorf("reading install source %q: %w", sourceDirectory, err)
	}
	syncs := make([]managedSync, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sync, planErr := planDirectorySync(filepath.Join(sourceDirectory, entry.Name()), filepath.Join(destinationDirectory, entry.Name()), false)
		if planErr != nil {
			return nil, planErr
		}
		syncs = append(syncs, sync)
	}
	return syncs, nil
}

func sameLexicalPath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}

func printHeading(output io.Writer, agent, source, agentHome string, apply, configure bool) {
	mode := "DRY RUN"
	if apply && configure {
		mode = "APPLY + CONFIGURE"
	} else if apply {
		mode = "APPLY"
	} else if configure {
		mode = "CONFIGURE REVIEW"
	}
	fmt.Fprintf(output, "SDLC installer: %s\n", mode)
	fmt.Fprintf(output, "Agent: %s\nSource: %s\nAgent home: %s\n", agent, source, agentHome)
}

func printInstallationPlan(output io.Writer, plan installationPlan, apply bool) {
	for _, sync := range plan.syncs {
		if !sync.needsSync {
			if os.Getenv("VERBOSE") == "1" {
				fmt.Fprintf(output, "Installation: %s already matches %s.\n", sync.destination, sync.source)
			}
			continue
		}
		verb := "would synchronize"
		if sync.backupExisting {
			verb = "would back up the existing path and synchronize"
		} else if sync.destinationExists {
			verb = "would back up drifted files and synchronize"
		}
		if apply {
			verb = "will synchronize"
			if sync.backupExisting {
				verb = "will back up the existing path and synchronize"
			} else if sync.destinationExists {
				verb = "will back up drifted files and synchronize"
			}
		}
		fmt.Fprintf(output, "Installation: %s %s -> %s\n", verb, sync.source, sync.destination)
	}
}

func applyInstallation(plan installationPlan, output io.Writer) error {
	epoch := time.Now().Unix()
	for _, sync := range plan.syncs {
		if !sync.needsSync {
			continue
		}
		if err := synchronizeDirectory(sync, epoch, output); err != nil {
			return err
		}
	}
	return nil
}

func planDirectorySync(source, destination string, excludeGit bool) (managedSync, error) {
	sync := managedSync{source: source, destination: destination, excludeGit: excludeGit}
	info, err := os.Lstat(destination)
	if errors.Is(err, os.ErrNotExist) {
		sync.needsSync = true
		return sync, nil
	}
	if err != nil {
		return managedSync{}, fmt.Errorf("inspecting live directory %q: %w", destination, err)
	}
	sync.destinationExists = true
	if info.Mode()&os.ModeSymlink != 0 {
		sync.needsSync = true
		sync.backupExisting = true
		return sync, nil
	}
	if !info.IsDir() {
		sync.needsSync = true
		sync.backupExisting = true
		return sync, nil
	}
	differs, compareErr := directoryDiffers(sync)
	if compareErr != nil {
		return managedSync{}, fmt.Errorf("comparing live directory %q: %w", destination, compareErr)
	}
	sync.needsSync = differs
	return sync, nil
}

func synchronizeDirectory(sync managedSync, epoch int64, output io.Writer) error {
	if info, err := os.Lstat(sync.destination); err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		if _, backupErr := backupArtifact(output, sync.destination, epoch); backupErr != nil {
			return backupErr
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(sync.destination, 0o700); err != nil {
		return fmt.Errorf("creating live directory %q: %w", sync.destination, err)
	}
	if err := backupDriftedEntries(output, sync, epoch); err != nil {
		return err
	}
	if err := rsyncDirectory(sync); err != nil {
		return err
	}
	return nil
}

func rsyncDirectory(sync managedSync) error {
	arguments := []string{"-a"}
	if sync.excludeGit {
		arguments = append(arguments, "--exclude=/.git/")
	}
	arguments = append(arguments, "--", sync.source+string(os.PathSeparator), sync.destination+string(os.PathSeparator))
	output, err := exec.Command("rsync", arguments...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("synchronizing %q to %q: %w: %s", sync.source, sync.destination, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func backupDriftedEntries(output io.Writer, sync managedSync, epoch int64) error {
	return filepath.Walk(sync.source, func(sourcePath string, sourceInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(sync.source, sourcePath)
		if err != nil || relative == "." {
			return err
		}
		if excludedFromSync(relative, sourceInfo, sync.excludeGit) {
			if sourceInfo.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		destinationPath := filepath.Join(sync.destination, relative)
		destinationInfo, err := os.Lstat(destinationPath)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if sourceInfo.IsDir() && destinationInfo.IsDir() && destinationInfo.Mode()&os.ModeSymlink == 0 {
			return nil
		}
		if sourceInfo.Mode().IsRegular() && destinationInfo.Mode().IsRegular() {
			equal, compareErr := regularFilesEqual(sourcePath, sourceInfo, destinationPath, destinationInfo)
			if compareErr != nil || equal {
				return compareErr
			}
		}
		_, err = backupArtifact(output, destinationPath, epoch)
		return err
	})
}

func directoryDiffers(sync managedSync) (bool, error) {
	differs := false
	err := filepath.Walk(sync.source, func(sourcePath string, sourceInfo os.FileInfo, walkErr error) error {
		if walkErr != nil || differs {
			return walkErr
		}
		relative, err := filepath.Rel(sync.source, sourcePath)
		if err != nil || relative == "." {
			return err
		}
		if excludedFromSync(relative, sourceInfo, sync.excludeGit) {
			if sourceInfo.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		destinationPath := filepath.Join(sync.destination, relative)
		destinationInfo, err := os.Lstat(destinationPath)
		if errors.Is(err, os.ErrNotExist) {
			differs = true
			return nil
		}
		if err != nil {
			return err
		}
		if sourceInfo.IsDir() {
			differs = !destinationInfo.IsDir() || destinationInfo.Mode()&os.ModeSymlink != 0
			return nil
		}
		if !sourceInfo.Mode().IsRegular() || !destinationInfo.Mode().IsRegular() {
			differs = true
			return nil
		}
		equal, compareErr := regularFilesEqual(sourcePath, sourceInfo, destinationPath, destinationInfo)
		if compareErr != nil {
			return compareErr
		}
		differs = !equal
		return nil
	})
	return differs, err
}

func excludedFromSync(relative string, info os.FileInfo, excludeGit bool) bool {
	return excludeGit && info.IsDir() && strings.Split(relative, string(os.PathSeparator))[0] == ".git"
}

func regularFilesEqual(sourcePath string, sourceInfo os.FileInfo, destinationPath string, destinationInfo os.FileInfo) (bool, error) {
	if sourceInfo.Size() != destinationInfo.Size() || sourceInfo.Mode().Perm() != destinationInfo.Mode().Perm() {
		return false, nil
	}
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return false, err
	}
	destinationContent, err := os.ReadFile(destinationPath)
	if err != nil {
		return false, err
	}
	return bytes.Equal(sourceContent, destinationContent), nil
}

func backupArtifact(output io.Writer, path string, epoch int64) (string, error) {
	backup := fmt.Sprintf("%s.%d.bak", path, epoch)
	if _, err := os.Lstat(backup); err == nil {
		return "", fmt.Errorf("backup destination already exists: %s", backup)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.Rename(path, backup); err != nil {
		return "", fmt.Errorf("backing up drifted artefact %q: %w", path, err)
	}
	fmt.Fprintf(output, "Backup: %s\n", backup)
	return backup, nil
}

func analyseConfiguration(agent, agentHome, source string, output io.Writer) (*configurationChange, error) {
	switch agent {
	case agentClaude:
		return analyseClaudeConfiguration(agentHome, output)
	case agentCodex:
		return analyseCodexConfiguration(agentHome, source, output)
	case agentHermes:
		return analyseHermesConfiguration(agentHome, source, output)
	case agentCopilot:
		fmt.Fprintln(output, "Configuration: Copilot-specific modification is not implemented; review provider instructions manually.")
		return nil, nil
	case agentCustom:
		fmt.Fprintln(output, "Configuration: custom target; no provider configuration assumptions were made.")
		return nil, nil
	default:
		return nil, fmt.Errorf("analysing unsupported agent %q", agent)
	}
}

func analyseClaudeConfiguration(agentHome string, output io.Writer) (*configurationChange, error) {
	path := filepath.Join(agentHome, "settings.json")
	settingsSymlink, err := pathIsSymlink(path)
	if err != nil {
		return nil, err
	}
	settings, original, mode, err := readClaudeSettings(path)
	if err != nil {
		return nil, err
	}
	changes, err := normalizeClaudeCommandRules(settings)
	if err != nil {
		return nil, fmt.Errorf("analysing %s: %w", path, err)
	}
	if len(changes.addedToDeny) == 0 && len(changes.removedFromAllow) == 0 {
		fmt.Fprintln(output, "Configuration: Claude settings already contain the SDLC command restrictions.")
		return nil, nil
	}
	for _, rule := range changes.addedToDeny {
		fmt.Fprintf(output, "Recommendation: add %s to permissions.deny.\n", rule)
	}
	for _, rule := range changes.removedFromAllow {
		fmt.Fprintf(output, "Recommendation: remove conflicting %s from permissions.allow.\n", rule)
	}
	if settingsSymlink {
		fmt.Fprintf(output, "Configuration: %s is a symlink; no automatic change will replace it. Edit its target manually.\n", path)
		return nil, nil
	}
	candidate, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding proposed %s: %w", path, err)
	}
	candidate = append(candidate, '\n')
	return &configurationChange{
		path:        path,
		beforeLabel: formatClaudeDenyBefore(original),
		afterLabel:  formatClaudePolicyAfter(changes),
		contents:    candidate,
		mode:        mode,
	}, nil
}

func pathIsSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspecting %s: %w", path, err)
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}

func readClaudeSettings(path string) (map[string]any, []byte, os.FileMode, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]any), nil, 0o600, nil
	}
	if err != nil {
		return nil, nil, 0, fmt.Errorf("reading %s: %w", path, err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, nil, 0, fmt.Errorf("parsing %s: %w", path, err)
	}
	if settings == nil {
		return nil, nil, 0, fmt.Errorf("parsing %s: top level must be a JSON object", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("inspecting %s: %w", path, err)
	}
	return settings, data, info.Mode().Perm(), nil
}

type claudePolicyChanges struct {
	addedToDeny      []string
	removedFromAllow []string
}

func normalizeClaudeCommandRules(settings map[string]any) (claudePolicyChanges, error) {
	permissions, err := objectField(settings, "permissions")
	if err != nil {
		return claudePolicyChanges{}, err
	}
	deny, err := stringSliceField(permissions, "deny")
	if err != nil {
		return claudePolicyChanges{}, err
	}
	allow, err := stringSliceField(permissions, "allow")
	if err != nil {
		return claudePolicyChanges{}, err
	}
	changes := claudePolicyChanges{
		addedToDeny:      make([]string, 0, len(claudeDeniedCommands)),
		removedFromAllow: make([]string, 0, len(claudeDeniedCommands)),
	}
	for _, rule := range claudeDeniedCommands {
		if !containsString(deny, rule) {
			deny = append(deny, rule)
			changes.addedToDeny = append(changes.addedToDeny, rule)
		}
	}
	filteredAllow := make([]string, 0, len(allow))
	for _, rule := range allow {
		if containsString(claudeDeniedCommands, rule) {
			changes.removedFromAllow = append(changes.removedFromAllow, rule)
			continue
		}
		filteredAllow = append(filteredAllow, rule)
	}
	permissions["deny"] = deny
	if len(changes.removedFromAllow) != 0 {
		permissions["allow"] = filteredAllow
	}
	settings["permissions"] = permissions
	return changes, nil
}

func objectField(parent map[string]any, key string) (map[string]any, error) {
	value, exists := parent[key]
	if !exists {
		return make(map[string]any), nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a JSON object", key)
	}
	return object, nil
}

func stringSliceField(parent map[string]any, key string) ([]string, error) {
	value, exists := parent[key]
	if !exists {
		return nil, nil
	}
	items, ok := value.([]any)
	if !ok {
		if strings, stringsOK := value.([]string); stringsOK {
			return strings, nil
		}
		return nil, fmt.Errorf("permissions.%s must be a JSON array", key)
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("permissions.%s entries must be strings", key)
		}
		result = append(result, text)
	}
	return result, nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func formatClaudeDenyBefore(original []byte) string {
	if len(original) == 0 {
		return "permissions.deny: absent settings.json"
	}
	return "permissions.deny: existing entries preserved; JSON spacing and key order may be normalized"
}

func formatClaudePolicyAfter(changes claudePolicyChanges) string {
	parts := make([]string, 0, 2)
	if len(changes.addedToDeny) != 0 {
		parts = append(parts, "permissions.deny adds: "+strings.Join(changes.addedToDeny, ", "))
	}
	if len(changes.removedFromAllow) != 0 {
		parts = append(parts, "permissions.allow removes conflicts: "+strings.Join(changes.removedFromAllow, ", "))
	}
	return strings.Join(parts, "; ")
}

type codexConfig struct {
	ApprovalPolicy     string `toml:"approval_policy"`
	SandboxMode        string `toml:"sandbox_mode"`
	ProjectDocMaxBytes int64  `toml:"project_doc_max_bytes"`
	AllowLoginShell    bool   `toml:"allow_login_shell"`
}

func analyseCodexConfiguration(agentHome, source string, output io.Writer) (*configurationChange, error) {
	configPath := filepath.Join(agentHome, "config.toml")
	config, metadata, exists, err := readCodexConfig(configPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		fmt.Fprintf(output, "Recommendation: %s is absent; review templates/codex-config.example.toml.\n", configPath)
	} else {
		printCodexRecommendations(output, config, metadata)
	}
	rulesPath := filepath.Join(agentHome, "rules", "sdlc.rules")
	templatePath := filepath.Join(source, "templates", "codex-sdlc.rules.example")
	contents, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("reading Codex rules template %s: %w", templatePath, err)
	}
	info, err := os.Lstat(rulesPath)
	if err == nil {
		if !info.Mode().IsRegular() {
			fmt.Fprintf(output, "Configuration: %s is not a regular file; cannot safely migrate it automatically.\n", rulesPath)
			return nil, nil
		}
		current, readErr := os.ReadFile(rulesPath)
		if readErr != nil {
			return nil, fmt.Errorf("reading %s: %w", rulesPath, readErr)
		}
		merged, changed, safe, mergeErr := mergeCodexRules(current, contents)
		if mergeErr != nil {
			return nil, fmt.Errorf("analysing %s: %w", rulesPath, mergeErr)
		}
		if !safe {
			fmt.Fprintf(output, "Configuration: existing %s cannot safely migrate automatically; compare it with the repository example manually.\n", rulesPath)
			return nil, nil
		}
		if !changed {
			fmt.Fprintln(output, "Configuration: Codex rules already contain the SDLC command restrictions.")
			return nil, nil
		}
		fmt.Fprintf(output, "Recommendation: migrate the managed SDLC command restrictions in %s.\n", rulesPath)
		return &configurationChange{
			path:        rulesPath,
			beforeLabel: "recognized existing SDLC rules; unrelated rules preserved",
			afterLabel:  "managed SDLC Python command restrictions updated",
			contents:    merged,
			mode:        info.Mode().Perm(),
		}, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("inspecting %s: %w", rulesPath, err)
	}
	fmt.Fprintf(output, "Recommendation: create %s from %s.\n", rulesPath, templatePath)
	return &configurationChange{
		path:        rulesPath,
		beforeLabel: "file absent",
		afterLabel:  string(contents),
		contents:    contents,
		mode:        0o600,
	}, nil
}

func mergeCodexRules(current, desired []byte) ([]byte, bool, bool, error) {
	if bytes.Equal(current, desired) {
		return current, false, true, nil
	}
	desiredBlock, err := markedBlock(desired, codexPythonRulesStart, codexPythonRulesEnd)
	if err != nil {
		return nil, false, false, fmt.Errorf("invalid repository rules template: %w", err)
	}
	startCount := bytes.Count(current, []byte(codexPythonRulesStart))
	endCount := bytes.Count(current, []byte(codexPythonRulesEnd))
	if startCount > 1 || endCount > 1 || startCount != endCount {
		return nil, false, false, nil
	}
	if startCount == 1 {
		currentBlock, blockErr := markedBlock(current, codexPythonRulesStart, codexPythonRulesEnd)
		if blockErr != nil {
			return nil, false, false, nil
		}
		merged := bytes.Replace(current, currentBlock, desiredBlock, 1)
		return merged, !bytes.Equal(current, merged), true, nil
	}
	for _, legacyRule := range []string{"rm", "sed", "awk", "source"} {
		if !bytes.Contains(current, []byte(`pattern = ["`+legacyRule+`"]`)) {
			return nil, false, false, nil
		}
	}
	if bytes.Contains(current, []byte(`pattern = ["python"]`)) || bytes.Contains(current, []byte(`pattern = ["python3"]`)) {
		return nil, false, false, nil
	}
	merged := append(bytes.TrimRight(current, "\n"), '\n', '\n')
	merged = append(merged, desiredBlock...)
	merged = append(merged, '\n')
	return merged, true, true, nil
}

func markedBlock(contents []byte, startMarker, endMarker string) ([]byte, error) {
	start := bytes.Index(contents, []byte(startMarker))
	end := bytes.Index(contents, []byte(endMarker))
	if start < 0 || end < start || bytes.Count(contents, []byte(startMarker)) != 1 || bytes.Count(contents, []byte(endMarker)) != 1 {
		return nil, errors.New("managed rule markers are missing or ambiguous")
	}
	end += len(endMarker)
	return contents[start:end], nil
}

func readCodexConfig(path string) (codexConfig, toml.MetaData, bool, error) {
	var config codexConfig
	metadata, err := toml.DecodeFile(path, &config)
	if errors.Is(err, os.ErrNotExist) {
		return config, metadata, false, nil
	}
	if err != nil {
		return config, metadata, false, fmt.Errorf("parsing %s: %w", path, err)
	}
	return config, metadata, true, nil
}

func printCodexRecommendations(output io.Writer, config codexConfig, metadata toml.MetaData) {
	if config.ApprovalPolicy == "never" {
		fmt.Fprintln(output, "Recommendation: approval_policy is never; review whether on-request better supports human checkpoints.")
	} else if !metadata.IsDefined("approval_policy") {
		fmt.Fprintln(output, "Recommendation: set an explicit approval_policy after reviewing current Codex documentation.")
	}
	if config.SandboxMode != "workspace-write" {
		fmt.Fprintln(output, "Recommendation: review sandbox_mode; the SDLC example uses workspace-write for local project work.")
	}
	if !metadata.IsDefined("allow_login_shell") || config.AllowLoginShell {
		fmt.Fprintln(output, "Recommendation: set allow_login_shell = false unless the project requires login-shell initialization.")
	}
	if config.ProjectDocMaxBytes < 65536 {
		fmt.Fprintln(output, "Recommendation: consider project_doc_max_bytes = 65536 for layered instruction documents.")
	}
}

func offerConfigurationChange(change *configurationChange, input io.Reader, output io.Writer) error {
	if change == nil {
		fmt.Fprintln(output, "Configuration: no automatic change is available for this target.")
		return nil
	}
	fmt.Fprintf(output, "Proposed configuration change: %s\n", change.path)
	fmt.Fprintf(output, "- before: %s\n+ after: %s\n", change.beforeLabel, change.afterLabel)
	fmt.Fprint(output, "Apply this configuration change? Type yes to continue: ")
	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading configuration confirmation: %w", err)
		}
		fmt.Fprintln(output, "Configuration unchanged.")
		return nil
	}
	if strings.ToLower(strings.TrimSpace(scanner.Text())) != "yes" {
		fmt.Fprintln(output, "Configuration unchanged.")
		return nil
	}
	backupPath, err := backupConfiguration(change.path)
	if err != nil {
		return err
	}
	if backupPath != "" {
		fmt.Fprintf(output, "Backup: %s\n", backupPath)
	}
	if err := writeFileAtomic(change.path, change.contents, change.mode); err != nil {
		if backupPath != "" {
			_ = os.Rename(backupPath, change.path)
		}
		return err
	}
	fmt.Fprintf(output, "Configuration updated: %s\n", change.path)
	return nil
}

func backupConfiguration(path string) (string, error) {
	_, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("inspecting configuration backup source %q: %w", path, err)
	}
	return backupArtifact(io.Discard, path, time.Now().Unix())
}

func writeFileAtomic(path string, contents []byte, mode os.FileMode) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("creating configuration directory %q: %w", directory, err)
	}
	temporary, err := os.CreateTemp(directory, ".sdlc-install-*")
	if err != nil {
		return fmt.Errorf("creating temporary configuration beside %q: %w", path, err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		// The file may already have been renamed into place; cleanup is best-effort.
		_ = os.Remove(temporaryPath)
	}()
	if err := temporary.Chmod(mode); err != nil {
		operationErr := fmt.Errorf("setting mode on temporary configuration %q: %w", temporaryPath, err)
		return closeAfterFailure(temporary, operationErr)
	}
	if _, err := io.Copy(temporary, bytes.NewReader(contents)); err != nil {
		operationErr := fmt.Errorf("writing temporary configuration %q: %w", temporaryPath, err)
		return closeAfterFailure(temporary, operationErr)
	}
	if err := temporary.Sync(); err != nil {
		operationErr := fmt.Errorf("syncing temporary configuration %q: %w", temporaryPath, err)
		return closeAfterFailure(temporary, operationErr)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing temporary configuration %q: %w", temporaryPath, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replacing configuration %q: %w", path, err)
	}
	return nil
}

func closeAfterFailure(file *os.File, operationErr error) error {
	if err := file.Close(); err != nil {
		return errors.Join(operationErr, fmt.Errorf("closing temporary configuration %q: %w", file.Name(), err))
	}
	return operationErr
}
