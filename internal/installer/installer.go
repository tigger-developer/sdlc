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
	name             string
	commandDirectory bool
	commandFiles     bool
	skillLinks       bool
}

var providerDefinitions = []providerDefinition{
	{name: agentClaude, commandFiles: true, skillLinks: true},
	{name: agentCodex, commandDirectory: true},
	{name: agentCopilot, commandDirectory: true, skillLinks: true},
	{name: agentHermes, skillLinks: true},
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
	backup      bool
}

type managedLink struct {
	source      string
	destination string
	legacy      []string
	needsLink   bool
	replaceLink bool
}

type managedSync struct {
	source      string
	destination string
	excludeGit  bool
	needsSync   bool
	replaceLink bool
}

type installationPlan struct {
	syncs []managedSync
	links []managedLink
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
		if err := applyInstallation(plan); err != nil {
			return err
		}
		verified, verifyErr := planInstallation(agent, source, agentHome)
		if verifyErr != nil {
			return fmt.Errorf("verifying installation: %w", verifyErr)
		}
		if installationHasChanges(verified) {
			return errors.New("installation verification still reports changes")
		}
		fmt.Fprintln(options.Output, "Installation: synchronized the live tree and installed all planned links.")
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
	commonHome := filepath.Join(userHome, ".agents")
	sharedPlan, err := planSharedInstallation(source, commonHome)
	if err != nil {
		return err
	}
	agents, err := detectedAgents(userHome)
	if err != nil {
		return err
	}
	type namedPlan struct {
		agent string
		plan  installationPlan
	}
	providerPlans := make([]namedPlan, 0, len(agents))
	for _, agent := range agents {
		agentHome := filepath.Join(userHome, "."+agent)
		plan, err := planProviderInstallation(agent, source, filepath.Join(commonHome, "sdlc"), agentHome)
		if err != nil {
			return err
		}
		providerPlans = append(providerPlans, namedPlan{agent: agent, plan: plan})
	}

	fmt.Fprintln(output, "SDLC installer: INTERACTIVE")
	if len(agents) == 0 {
		fmt.Fprintln(output, "Detected agents: none")
	} else {
		fmt.Fprintf(output, "Detected agents: %s\n", strings.Join(agents, ", "))
	}
	reader := bufio.NewReader(input)
	if installationHasChanges(sharedPlan) {
		printInstallationPlan(output, sharedPlan, false)
		accepted, confirmErr := confirm(reader, output, "Deploy shared SDLC tree? [yes/no]: ")
		if confirmErr != nil {
			return confirmErr
		}
		if accepted {
			if err := applyInstallation(sharedPlan); err != nil {
				return err
			}
			verified, verifyErr := planSharedInstallation(source, commonHome)
			if verifyErr != nil {
				return verifyErr
			}
			if installationHasChanges(verified) {
				return errors.New("shared deployment verification still reports changes")
			}
			fmt.Fprintln(output, "Shared deployment installed.")
		} else {
			fmt.Fprintln(output, "Shared deployment declined.")
		}
	} else {
		fmt.Fprintln(output, "Shared deployment is current.")
	}

	for _, provider := range providerPlans {
		agent := provider.agent
		plan := provider.plan
		if !installationHasChanges(plan) {
			fmt.Fprintf(output, "Provider %s adapters are current.\n", agent)
			continue
		}
		printInstallationPlan(output, plan, false)
		accepted, confirmErr := confirm(reader, output, fmt.Sprintf("Install %s adapters? [yes/no]: ", agent))
		if confirmErr != nil {
			return confirmErr
		}
		if !accepted {
			fmt.Fprintf(output, "Provider %s adapters declined.\n", agent)
			continue
		}
		if err := applyInstallation(plan); err != nil {
			return err
		}
		verified, verifyErr := planProviderInstallation(agent, source, filepath.Join(commonHome, "sdlc"), filepath.Join(userHome, "."+agent))
		if verifyErr != nil {
			return verifyErr
		}
		if installationHasChanges(verified) {
			return fmt.Errorf("%s adapter verification still reports changes", agent)
		}
		fmt.Fprintf(output, "Provider %s adapters installed.\n", agent)
	}
	return nil
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
	for _, link := range plan.links {
		if link.needsLink {
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
	liveSDLC := filepath.Join(commonHome, "sdlc")
	if filepath.Clean(source) == filepath.Clean(filepath.Join(agentHome, "sdlc")) {
		return installationPlan{}, fmt.Errorf("source %q is inside the provider home; move the staging clone outside the provider home before installing the live adapter", source)
	}
	shared, err := planSharedInstallation(source, commonHome)
	if err != nil {
		return installationPlan{}, err
	}
	provider, err := planProviderInstallation(agent, source, liveSDLC, agentHome)
	if err != nil {
		return installationPlan{}, err
	}
	shared.links = append(shared.links, provider.links...)
	return shared, nil
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
	links, err := childLinks(filepath.Join(source, "skills"), filepath.Join(liveSDLC, "skills"), filepath.Join(commonHome, "skills"), true)
	if err != nil {
		return installationPlan{}, err
	}
	plan := installationPlan{syncs: []managedSync{sync}}
	planned, err := planLinks(links)
	if err != nil {
		return installationPlan{}, err
	}
	plan.links = planned
	return plan, nil
}

func planProviderInstallation(agent, source, liveSDLC, agentHome string) (installationPlan, error) {
	if filepath.Clean(source) == filepath.Clean(filepath.Join(agentHome, "sdlc")) {
		return installationPlan{}, fmt.Errorf("source %q is inside the provider home; move the staging clone outside the provider home before installing the live adapter", source)
	}
	links, err := providerLinks(agent, source, liveSDLC, agentHome)
	if err != nil {
		return installationPlan{}, err
	}
	planned, err := planLinks(links)
	if err != nil {
		return installationPlan{}, err
	}
	return installationPlan{links: planned}, nil
}

func planLinks(links []managedLink) ([]managedLink, error) {
	planned := make([]managedLink, 0, len(links))
	for _, link := range links {
		item, planErr := planLink(link)
		if planErr != nil {
			return nil, planErr
		}
		planned = append(planned, item)
	}
	return planned, nil
}

func providerLinks(agent, source, liveSDLC, agentHome string) ([]managedLink, error) {
	commonSkills := filepath.Join(filepath.Dir(agentHome), ".agents", "skills")
	links := []managedLink{{
		source:      liveSDLC,
		destination: filepath.Join(agentHome, "sdlc"),
		legacy:      []string{source},
	}}
	if agent == agentCustom {
		return links, nil
	}
	provider, found := providerDefinitionFor(agent)
	if !found {
		return nil, fmt.Errorf("planning unsupported agent %q", agent)
	}
	if provider.commandFiles {
		commandLinks, err := childLinks(filepath.Join(source, "commands"), filepath.Join(liveSDLC, "commands"), filepath.Join(agentHome, "commands"), false)
		if err != nil {
			return nil, err
		}
		links = append(links, commandLinks...)
	}
	if provider.commandDirectory {
		links = append(links, managedLink{
			source:      filepath.Join(liveSDLC, "commands"),
			destination: filepath.Join(agentHome, "prompts-commands"),
			legacy:      []string{filepath.Join(source, "commands")},
		})
	}
	if provider.skillLinks {
		providerSkills, err := childLinks(filepath.Join(source, "skills"), commonSkills, filepath.Join(agentHome, "skills"), true)
		if err != nil {
			return nil, err
		}
		links = append(links, providerSkills...)
	}
	return links, nil
}

func childLinks(stagingDirectory, sourceDirectory, destinationDirectory string, directoriesOnly bool) ([]managedLink, error) {
	entries, err := os.ReadDir(stagingDirectory)
	if err != nil {
		return nil, fmt.Errorf("reading install source %q: %w", stagingDirectory, err)
	}
	links := make([]managedLink, 0, len(entries))
	for _, entry := range entries {
		if directoriesOnly && !entry.IsDir() {
			continue
		}
		links = append(links, managedLink{
			source:      filepath.Join(sourceDirectory, entry.Name()),
			destination: filepath.Join(destinationDirectory, entry.Name()),
			legacy:      []string{filepath.Join(stagingDirectory, entry.Name())},
		})
	}
	return links, nil
}

func planLink(link managedLink) (managedLink, error) {
	if sameLexicalPath(link.source, link.destination) {
		return managedLink{}, fmt.Errorf("managed link source and destination are the same path %q", link.source)
	}
	info, err := os.Lstat(link.destination)
	if errors.Is(err, os.ErrNotExist) {
		link.needsLink = true
		return link, nil
	}
	if err != nil {
		return managedLink{}, fmt.Errorf("inspecting destination %q: %w", link.destination, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, readErr := os.Readlink(link.destination)
		if readErr != nil {
			return managedLink{}, fmt.Errorf("reading destination symlink %q: %w", link.destination, readErr)
		}
		if linkTargetMatches(link.destination, target, link.source) {
			return link, nil
		}
		for _, legacy := range link.legacy {
			if linkTargetMatches(link.destination, target, legacy) {
				link.needsLink = true
				link.replaceLink = true
				return link, nil
			}
		}
	}
	return managedLink{}, fmt.Errorf("destination %q already exists and does not point to %q", link.destination, link.source)
}

func linkTargetMatches(linkPath, target, wanted string) bool {
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(linkPath), target)
	}
	return sameLexicalPath(target, wanted)
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
			fmt.Fprintf(output, "Installation: %s already matches %s.\n", sync.destination, sync.source)
			continue
		}
		verb := "would synchronize"
		if apply {
			verb = "will synchronize"
		}
		fmt.Fprintf(output, "Installation: %s %s -> %s\n", verb, sync.source, sync.destination)
	}
	for _, link := range plan.links {
		if !link.needsLink {
			fmt.Fprintf(output, "Installation: %s already resolves to %s.\n", link.destination, link.source)
			continue
		}
		verb := "would create symlink"
		if apply {
			verb = "will create symlink"
		}
		fmt.Fprintf(output, "Installation: %s %s -> %s\n", verb, link.destination, link.source)
	}
}

func applyInstallation(plan installationPlan) error {
	for _, sync := range plan.syncs {
		if !sync.needsSync {
			continue
		}
		if err := synchronizeDirectory(sync); err != nil {
			return err
		}
	}
	for _, link := range plan.links {
		if !link.needsLink {
			continue
		}
		directory := filepath.Dir(link.destination)
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return fmt.Errorf("creating install directory %q: %w", directory, err)
		}
		if link.replaceLink {
			if err := os.Remove(link.destination); err != nil {
				return fmt.Errorf("removing recognized legacy symlink %q: %w", link.destination, err)
			}
		}
		if err := os.Symlink(link.source, link.destination); err != nil {
			return fmt.Errorf("creating symlink %q: %w", link.destination, err)
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
	if info.Mode()&os.ModeSymlink != 0 {
		target, readErr := os.Readlink(destination)
		if readErr != nil {
			return managedSync{}, fmt.Errorf("reading live directory symlink %q: %w", destination, readErr)
		}
		if !linkTargetMatches(destination, target, source) {
			return managedSync{}, fmt.Errorf("live directory %q points to unrelated target %q", destination, target)
		}
		sync.needsSync = true
		sync.replaceLink = true
		return sync, nil
	}
	if !info.IsDir() {
		return managedSync{}, fmt.Errorf("live directory %q exists and is not a directory", destination)
	}
	changes, compareErr := rsyncDirectory(sync, true)
	if compareErr != nil {
		return managedSync{}, fmt.Errorf("comparing live directory %q: %w", destination, compareErr)
	}
	sync.needsSync = len(bytes.TrimSpace(changes)) != 0
	return sync, nil
}

func synchronizeDirectory(sync managedSync) error {
	if sync.replaceLink {
		if err := os.Remove(sync.destination); err != nil {
			return fmt.Errorf("removing recognized staging symlink %q: %w", sync.destination, err)
		}
	}
	if err := os.MkdirAll(sync.destination, 0o700); err != nil {
		return fmt.Errorf("creating live directory %q: %w", sync.destination, err)
	}
	if _, err := rsyncDirectory(sync, false); err != nil {
		return err
	}
	changes, err := rsyncDirectory(sync, true)
	if err != nil {
		return fmt.Errorf("verifying live directory %q: %w", sync.destination, err)
	}
	if len(bytes.TrimSpace(changes)) != 0 {
		return fmt.Errorf("verifying live directory %q: rsync still reports changes: %s", sync.destination, strings.TrimSpace(string(changes)))
	}
	return nil
}

func rsyncDirectory(sync managedSync, dryRun bool) ([]byte, error) {
	arguments := []string{"-a", "--delete"}
	if sync.excludeGit {
		arguments = append(arguments, "--exclude=/.git", "--delete-excluded")
	}
	if dryRun {
		arguments = append(arguments, "--dry-run", "--itemize-changes")
	}
	arguments = append(arguments, "--", sync.source+string(os.PathSeparator), sync.destination+string(os.PathSeparator))
	output, err := exec.Command("rsync", arguments...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("synchronizing %q to %q: %w: %s", sync.source, sync.destination, err, strings.TrimSpace(string(output)))
	}
	return output, nil
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
	if change.backup {
		backupPath, err := backupConfiguration(change.path)
		if err != nil {
			return err
		}
		if backupPath != "" {
			fmt.Fprintf(output, "Backup: %s\n", backupPath)
		}
	}
	if err := writeFileAtomic(change.path, change.contents, change.mode); err != nil {
		return err
	}
	fmt.Fprintf(output, "Configuration updated: %s\n", change.path)
	return nil
}

func backupConfiguration(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading configuration backup source %q: %w", path, err)
	}
	directory, err := os.MkdirTemp("", "sdlc-install-hermes-config-")
	if err != nil {
		return "", fmt.Errorf("creating Hermes configuration backup directory: %w", err)
	}
	backupPath := filepath.Join(directory, filepath.Base(path))
	if err := os.WriteFile(backupPath, contents, 0o600); err != nil {
		return "", fmt.Errorf("writing Hermes configuration backup %q: %w", backupPath, err)
	}
	return backupPath, nil
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
