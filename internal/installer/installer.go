package installer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
}

var providerDefinitions = []providerDefinition{
	{name: agentClaude, commandPath: "commands"},
	{name: agentCodex, commandPath: "prompts-commands"},
	{name: agentCopilot, commandPath: "prompts-commands"},
	{name: agentHermes},
}

var retiredCommandFiles = []string{
	"build.md",
	"end-discovery.md",
	"implement.md",
	"migrate-acs.md",
	"review.md",
	"start-discovery.md",
	"write-tests.md",
}

var retiredSkillFiles = []string{
	filepath.Join("audit-acs", "SKILL.md"),
	filepath.Join("design-solution", "SKILL.md"),
	filepath.Join("draft-bug-fix", "SKILL.md"),
	filepath.Join("draft-design-issue", "SKILL.md"),
	filepath.Join("draft-issue", "SKILL.md"),
}

var retiredProviderArtifacts = []string{
	"audit-code",
	"audit-design",
	"audit-spec",
	"audit-tests",
	"convert-migrated-acs-to-org",
	"diagnose-issue",
	"migrate-legacy-acs-to-sdlc-v1",
	"recommendations-please",
	"summarize-issues",
	"useful-be",
}

var retiredSharedFiles = []string{
	"GO.md",
	"PERL.md",
	"PYTHON.md",
	"SHELL.md",
	"SWIFT.md",
	"WEB.md",
	filepath.Join("presets", "sdlc-standards", "templates", "constitution-addendum.md"),
}

var retiredV2SharedPaths = []string{
	"commands",
	"skills",
	filepath.Join("presets", "sdlc-standards"),
	"prompts",
	filepath.Join("templates", "project-init"),
}

var retiredGlobalSkillPaths = []string{
	"audit-acs",
	"design-solution",
	"draft-bug-fix",
	"draft-design-issue",
	"draft-issue",
}

const (
	codexPythonRulesStart = "# BEGIN SDLC MANAGED PYTHON RULES"
	codexPythonRulesEnd   = "# END SDLC MANAGED PYTHON RULES"
	toolGuardCommand      = "bash ~/.agents/sdlc/hooks/agent-command-guard.sh"
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
	needsSync         bool
	backupExisting    bool
	destinationExists bool
}

type managedRetirement struct {
	path string
}

type installationPlan struct {
	syncs          []managedSync
	retirements    []managedRetirement
	configurations []*configurationChange
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
	changes, err := analyseConfiguration(agent, agentHome, source, options.Output)
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
		if err := verifyCanonicalMain(filepath.Join(filepath.Dir(agentHome), ".agents")); err != nil {
			return err
		}
		fmt.Fprintln(options.Output, "Installation: synchronized all planned copies.")
	}
	if options.Configure {
		return offerConfigurationChanges(changes, options.Input, options.Output)
	}
	if len(changes) != 0 {
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
	shared, planErr := planSharedInstallation(source, commonHome)
	if planErr != nil {
		return planErr
	}
	mergeInstallationPlan(&plan, shared)
	for _, agent := range agents {
		agentHome := filepath.Join(userHome, "."+agent)
		provider, planErr := planProviderInstallation(agent, source, agentHome)
		if planErr != nil {
			return planErr
		}
		mergeInstallationPlan(&plan, provider)
	}
	configurations, err := planDetectedConfigurations(source, userHome, agents)
	if err != nil {
		return err
	}
	plan.configurations = configurations

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
	for _, change := range plan.configurations {
		printConfigurationChange(output, change)
	}
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
	for _, change := range plan.configurations {
		if err := applyConfigurationChange(change, output); err != nil {
			return err
		}
	}
	verified, verifyErr := planDetectedInstallation(source, userHome, agents)
	if verifyErr != nil {
		return verifyErr
	}
	if installationHasChanges(verified) {
		return errors.New("deployment verification still reports changes")
	}
	if err := verifyCanonicalMain(commonHome); err != nil {
		return err
	}
	fmt.Fprintln(output, "All listed SDLC changes were installed.")
	return nil
}

func planDetectedInstallation(source, userHome string, agents []string) (installationPlan, error) {
	plan := installationPlan{}
	commonHome := filepath.Join(userHome, ".agents")
	shared, err := planSharedInstallation(source, commonHome)
	if err != nil {
		return installationPlan{}, err
	}
	mergeInstallationPlan(&plan, shared)
	for _, agent := range agents {
		provider, err := planProviderInstallation(agent, source, filepath.Join(userHome, "."+agent))
		if err != nil {
			return installationPlan{}, err
		}
		mergeInstallationPlan(&plan, provider)
	}
	configurations, err := planDetectedConfigurations(source, userHome, agents)
	if err != nil {
		return installationPlan{}, err
	}
	plan.configurations = configurations
	return plan, nil
}

func planDetectedConfigurations(source, userHome string, agents []string) ([]*configurationChange, error) {
	var changes []*configurationChange
	for _, agent := range agents {
		agentChanges, err := analyseConfiguration(agent, filepath.Join(userHome, "."+agent), source, io.Discard)
		if err != nil {
			return nil, err
		}
		changes = append(changes, agentChanges...)
	}
	return changes, nil
}

func confirm(reader *bufio.Reader, output io.Writer, prompt string) (bool, error) {
	fmt.Fprint(output, prompt)
	response, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("reading installation confirmation: %w", err)
	}
	return confirmationAccepted(response), nil
}

func confirmationAccepted(response string) bool {
	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		return true
	default:
		return false
	}
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
	if len(plan.configurations) != 0 || len(plan.retirements) != 0 {
		return true
	}
	for _, sync := range plan.syncs {
		if sync.needsSync {
			return true
		}
	}
	return false
}

func mergeInstallationPlan(plan *installationPlan, addition installationPlan) {
	plan.syncs = append(plan.syncs, addition.syncs...)
	plan.retirements = append(plan.retirements, addition.retirements...)
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
	for _, required := range []string{filepath.Join("src", "MAIN.md"), "README.md"} {
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
	plan := installationPlan{}
	shared, err := planSharedInstallation(source, commonHome)
	if err != nil {
		return installationPlan{}, err
	}
	mergeInstallationPlan(&plan, shared)
	provider, err := planProviderInstallation(agent, source, agentHome)
	if err != nil {
		return installationPlan{}, err
	}
	mergeInstallationPlan(&plan, provider)
	return plan, nil
}

func planSharedInstallation(source, commonHome string) (installationPlan, error) {
	liveSDLC := filepath.Join(commonHome, "sdlc")
	if sameLexicalPath(source, liveSDLC) {
		return installationPlan{}, fmt.Errorf("source %q is the live SDLC directory; use a separate staging clone", source)
	}
	plan := installationPlan{}
	for _, mapping := range []struct {
		source, destination string
		optional            bool
	}{
		{source: filepath.Join(source, "src"), destination: liveSDLC},
		{source: filepath.Join(source, "commands"), destination: filepath.Join(liveSDLC, "commands"), optional: true},
		{source: filepath.Join(source, "hooks"), destination: filepath.Join(liveSDLC, "hooks")},
		{source: filepath.Join(source, "skills"), destination: filepath.Join(commonHome, "skills")},
	} {
		files, err := planDirectoryMapping(mapping.source, mapping.destination, mapping.optional)
		if err != nil {
			return installationPlan{}, err
		}
		plan.syncs = append(plan.syncs, files...)
	}
	for _, retirement := range []struct {
		sourceRoot, destinationRoot string
		files                       []string
	}{
		{filepath.Join(source, "src"), liveSDLC, retiredSharedFiles},
		{filepath.Join(source, "commands"), filepath.Join(liveSDLC, "commands"), retiredCommandFiles},
		{filepath.Join(source, "skills"), filepath.Join(commonHome, "skills"), retiredSkillFiles},
	} {
		retirements, err := planRetiredFiles(retirement.sourceRoot, retirement.destinationRoot, retirement.files)
		if err != nil {
			return installationPlan{}, err
		}
		plan.retirements = append(plan.retirements, retirements...)
	}
	sharedRetirements, err := planExactRetirements(liveSDLC, retiredV2SharedPaths)
	if err != nil {
		return installationPlan{}, err
	}
	plan.retirements = append(plan.retirements, sharedRetirements...)
	retirements, retirementErr := planExactRetirements(filepath.Join(commonHome, "skills"), retiredGlobalSkillPaths)
	if retirementErr != nil {
		return installationPlan{}, retirementErr
	}
	plan.retirements = append(plan.retirements, retirements...)
	return plan, nil
}

func verifyCanonicalMain(commonHome string) error {
	mainPath := filepath.Join(commonHome, "sdlc", "MAIN.md")
	info, err := os.Stat(mainPath)
	if err != nil {
		return fmt.Errorf("verifying canonical SDLC entry point %q: %w", mainPath, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("verifying canonical SDLC entry point %q: expected a regular file", mainPath)
	}
	return nil
}

func planProviderInstallation(agent, source, agentHome string) (installationPlan, error) {
	plan := installationPlan{}
	if agent == agentCustom {
		return plan, nil
	}
	provider, found := providerDefinitionFor(agent)
	if !found {
		return installationPlan{}, fmt.Errorf("planning unsupported agent %q", agent)
	}
	if agent == agentCodex && provider.commandPath != "" {
		sourceRoot := filepath.Join(source, "commands")
		destinationRoot := filepath.Join(agentHome, provider.commandPath)
		retirements, planErr := planRetiredFiles(sourceRoot, destinationRoot, retiredCommandFiles)
		if planErr != nil {
			return installationPlan{}, planErr
		}
		plan.retirements = append(plan.retirements, retirements...)
	}
	if agent != agentCodex {
		paths := append([]string{}, retiredProviderArtifacts...)
		currentSkills, planErr := sourceDirectoryNames(filepath.Join(source, "skills"))
		if planErr != nil {
			return installationPlan{}, planErr
		}
		paths = append(paths, currentSkills...)
		for _, relative := range retiredSkillFiles {
			paths = append(paths, filepath.Dir(relative))
		}
		retirements, planErr := planExactRetirements(filepath.Join(agentHome, "skills"), paths)
		if planErr != nil {
			return installationPlan{}, planErr
		}
		plan.retirements = append(plan.retirements, retirements...)
		if provider.commandPath != "" {
			commandRetirements, commandErr := planExactRetirements(filepath.Join(agentHome, provider.commandPath), retiredCommandFiles)
			if commandErr != nil {
				return installationPlan{}, commandErr
			}
			plan.retirements = append(plan.retirements, commandRetirements...)
		}
		if agent == agentCopilot {
			hookRetirements, hookErr := planExactRetirements(agentHome, []string{filepath.Join("hooks", "sdlc-tool-guard.json")})
			if hookErr != nil {
				return installationPlan{}, hookErr
			}
			plan.retirements = append(plan.retirements, hookRetirements...)
		}
	}
	return plan, nil
}

func sourceDirectoryNames(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("reading managed source directories %q: %w", root, err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func planExactRetirements(root string, relativePaths []string) ([]managedRetirement, error) {
	seen := map[string]bool{}
	var retirements []managedRetirement
	for _, relative := range relativePaths {
		path := filepath.Join(root, relative)
		if seen[path] {
			continue
		}
		seen[path] = true
		if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("inspecting retired provider artefact %q: %w", path, err)
		}
		retirements = append(retirements, managedRetirement{path: path})
	}
	return retirements, nil
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
				fmt.Fprintf(output, "Installation: current %s <- %s\n", sync.destination, sync.source)
			}
			continue
		}
		verb := "would install missing file"
		if sync.backupExisting {
			verb = "would back up and replace differing file"
		} else if sync.destinationExists {
			verb = "would back up and replace differing file"
		}
		if apply {
			verb = "will install missing file"
			if sync.backupExisting {
				verb = "will back up and replace differing file"
			} else if sync.destinationExists {
				verb = "will back up and replace differing file"
			}
		}
		fmt.Fprintf(output, "Installation: %s %s <- %s\n", verb, sync.destination, sync.source)
	}
	for _, retirement := range plan.retirements {
		verb := "would back up and retire legacy SDLC artefact"
		if apply {
			verb = "will back up and retire legacy SDLC artefact"
		}
		fmt.Fprintf(output, "Installation: %s %s\n", verb, retirement.path)
	}
}

func applyInstallation(plan installationPlan, output io.Writer) error {
	epoch := time.Now().Unix()
	for _, sync := range plan.syncs {
		if !sync.needsSync {
			continue
		}
		if err := synchronizeFile(sync, epoch, output); err != nil {
			return err
		}
	}
	for _, retirement := range plan.retirements {
		if _, err := backupArtifact(output, retirement.path, epoch); err != nil {
			return err
		}
		fmt.Fprintf(output, "Installation retired: %s\n", retirement.path)
	}
	return nil
}

func planDirectoryMapping(sourceRoot, destinationRoot string, optional bool) ([]managedSync, error) {
	if optional {
		_, err := os.Stat(sourceRoot)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("inspecting optional install source %q: %w", sourceRoot, err)
		}
	}
	return planDirectoryFiles(sourceRoot, destinationRoot)
}

func planRetiredFiles(sourceRoot, destinationRoot string, relativePaths []string) ([]managedRetirement, error) {
	var retirements []managedRetirement
	for _, relativePath := range relativePaths {
		sourcePath := filepath.Join(sourceRoot, relativePath)
		if _, err := os.Lstat(sourcePath); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("inspecting retired install source %q: %w", sourcePath, err)
		}

		destinationPath := filepath.Join(destinationRoot, relativePath)
		obstructed, err := destinationParentObstructed(destinationPath, destinationRoot)
		if err != nil {
			return nil, err
		}
		if obstructed {
			continue
		}
		if _, err := os.Lstat(destinationPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("inspecting legacy SDLC artefact %q: %w", destinationPath, err)
		}
		retirements = append(retirements, managedRetirement{path: destinationPath})
	}
	return retirements, nil
}

func planDirectoryFiles(sourceRoot, destinationRoot string) ([]managedSync, error) {
	rootInfo, err := os.Stat(sourceRoot)
	if err != nil {
		return nil, fmt.Errorf("inspecting install source %q: %w", sourceRoot, err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("install source %q is not a directory", sourceRoot)
	}
	var syncs []managedSync
	err = filepath.Walk(sourceRoot, func(sourcePath string, sourceInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if sourceInfo.IsDir() {
			return nil
		}
		if !sourceInfo.Mode().IsRegular() {
			return fmt.Errorf("install source %q is not a regular file", sourcePath)
		}
		relative, err := filepath.Rel(sourceRoot, sourcePath)
		if err != nil {
			return err
		}
		sync, err := planFileSync(sourcePath, filepath.Join(destinationRoot, relative), destinationRoot)
		if err != nil {
			return err
		}
		syncs = append(syncs, sync)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discovering install source %q: %w", sourceRoot, err)
	}
	return syncs, nil
}

func planFileSync(source, destination, destinationRoot string) (managedSync, error) {
	sync := managedSync{source: source, destination: destination}
	obstructed, err := destinationParentObstructed(destination, destinationRoot)
	if err != nil {
		return managedSync{}, err
	}
	if obstructed {
		sync.needsSync = true
		return sync, nil
	}
	info, err := os.Lstat(destination)
	if errors.Is(err, os.ErrNotExist) {
		sync.needsSync = true
		return sync, nil
	}
	if err != nil {
		return managedSync{}, fmt.Errorf("inspecting live file %q: %w", destination, err)
	}
	sync.destinationExists = true
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		sync.needsSync = true
		sync.backupExisting = true
		return sync, nil
	}
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return managedSync{}, fmt.Errorf("inspecting install source %q: %w", source, err)
	}
	equal, err := regularFilesEqual(source, sourceInfo, destination, info)
	if err != nil {
		return managedSync{}, fmt.Errorf("comparing live file %q: %w", destination, err)
	}
	sync.needsSync = !equal
	return sync, nil
}

func destinationParentObstructed(destination, destinationRoot string) (bool, error) {
	var unresolved error
	reachedRoot := false
	for directory := filepath.Dir(destination); ; directory = filepath.Dir(directory) {
		info, err := os.Lstat(directory)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return true, nil
			}
			if reachedRoot {
				return false, unresolved
			}
		} else if !errors.Is(err, os.ErrNotExist) && unresolved == nil {
			unresolved = fmt.Errorf("inspecting destination parent %q: %w", directory, err)
		}
		if sameLexicalPath(directory, destinationRoot) {
			reachedRoot = true
			if unresolved != nil {
				continue
			}
			return false, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			if unresolved != nil {
				return false, unresolved
			}
			return false, nil
		}
	}
}

func synchronizeFile(sync managedSync, epoch int64, output io.Writer) error {
	if err := ensureRegularDirectory(filepath.Dir(sync.destination), epoch, output); err != nil {
		return err
	}
	if _, err := os.Lstat(sync.destination); err == nil {
		if _, err := backupArtifact(output, sync.destination, epoch); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspecting live file %q: %w", sync.destination, err)
	}
	contents, err := os.ReadFile(sync.source)
	if err != nil {
		return fmt.Errorf("reading install source %q: %w", sync.source, err)
	}
	info, err := os.Stat(sync.source)
	if err != nil {
		return fmt.Errorf("inspecting install source %q: %w", sync.source, err)
	}
	if err := writeFileAtomic(sync.destination, contents, info.Mode().Perm()); err != nil {
		return err
	}
	fmt.Fprintf(output, "Installation updated: %s\n", sync.destination)
	return nil
}

func ensureRegularDirectory(directory string, epoch int64, output io.Writer) error {
	info, err := os.Lstat(directory)
	if err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		parent := filepath.Dir(directory)
		if parent != directory {
			if err := ensureRegularDirectory(parent, epoch, output); err != nil {
				return err
			}
		}
		return os.Mkdir(directory, 0o700)
	}
	if err != nil {
		return fmt.Errorf("inspecting installation directory %q: %w", directory, err)
	}
	if _, err := backupArtifact(output, directory, epoch); err != nil {
		return err
	}
	return os.Mkdir(directory, 0o700)
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

func analyseConfiguration(agent, agentHome, source string, output io.Writer) ([]*configurationChange, error) {
	switch agent {
	case agentClaude:
		change, err := analyseClaudeRetirement(agentHome, output)
		return configurationChanges(change), err
	case agentCodex:
		return analyseCodexConfigurations(agentHome, source, output)
	case agentHermes:
		change, err := analyseHermesRetirement(agentHome, output)
		return configurationChanges(change), err
	case agentCopilot:
		return nil, nil
	case agentCustom:
		fmt.Fprintln(output, "Configuration: custom target; no provider configuration assumptions were made.")
		return nil, nil
	default:
		return nil, fmt.Errorf("analysing unsupported agent %q", agent)
	}
}

func configurationChanges(change *configurationChange) []*configurationChange {
	if change == nil {
		return nil
	}
	return []*configurationChange{change}
}

func analyseClaudeRetirement(agentHome string, output io.Writer) (*configurationChange, error) {
	path := filepath.Join(agentHome, "settings.json")
	settingsSymlink, err := pathIsSymlink(path)
	if err != nil {
		return nil, err
	}
	settings, original, mode, err := readClaudeSettings(path)
	if err != nil {
		return nil, err
	}
	if len(original) == 0 {
		return nil, nil
	}
	changed, err := removeClaudeToolGuard(settings)
	if err != nil {
		return nil, fmt.Errorf("analysing %s: %w", path, err)
	}
	if !changed {
		return nil, nil
	}
	if settingsSymlink {
		fmt.Fprintf(output, "Configuration: %s is a symlink; remove its SDLC tool guard manually.\n", path)
		return nil, nil
	}
	candidate, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding proposed %s: %w", path, err)
	}
	candidate = append(candidate, '\n')
	return &configurationChange{
		path: path, beforeLabel: "existing configuration; unrelated values preserved",
		afterLabel: "SDLC v1/v2 tool guard removed", contents: candidate, mode: mode,
	}, nil
}

func removeClaudeToolGuard(settings map[string]any) (bool, error) {
	value, exists := settings["hooks"]
	if !exists {
		return false, nil
	}
	hooks, ok := value.(map[string]any)
	if !ok {
		return false, errors.New("hooks must be a JSON object")
	}
	value, exists = hooks["PreToolUse"]
	if !exists {
		return false, nil
	}
	entries, ok := value.([]any)
	if !ok {
		return false, errors.New("hooks.PreToolUse must be a JSON array")
	}
	changed := false
	filteredGroups := make([]any, 0, len(entries))
	for _, item := range entries {
		group, ok := item.(map[string]any)
		if !ok {
			return false, errors.New("hooks.PreToolUse entries must be JSON objects")
		}
		handlerValue, ok := group["hooks"].([]any)
		if !ok {
			filteredGroups = append(filteredGroups, item)
			continue
		}
		filteredHandlers := make([]any, 0, len(handlerValue))
		for _, raw := range handlerValue {
			handler, ok := raw.(map[string]any)
			command, commandOK := handler["command"].(string)
			if ok && commandOK && isManagedGuardCommand(command) {
				changed = true
				continue
			}
			filteredHandlers = append(filteredHandlers, raw)
		}
		if len(filteredHandlers) != 0 {
			group["hooks"] = filteredHandlers
			filteredGroups = append(filteredGroups, group)
		}
	}
	if !changed {
		return false, nil
	}
	if len(filteredGroups) == 0 {
		delete(hooks, "PreToolUse")
	} else {
		hooks["PreToolUse"] = filteredGroups
	}
	if len(hooks) == 0 {
		delete(settings, "hooks")
	}
	return true, nil
}

func isManagedGuardCommand(command string) bool {
	command = strings.TrimSpace(command)
	if command == toolGuardCommand {
		return true
	}
	unquoted := strings.Trim(command, "\"")
	return strings.HasPrefix(unquoted, "bash ") &&
		(strings.Contains(unquoted, "/.agents/sdlc/hooks/agent-command-guard.sh") ||
			strings.Contains(unquoted, "/.hermes/sdlc/hooks/agent-command-guard.sh") ||
			strings.Contains(unquoted, "/.claude/sdlc/hooks/agent-command-guard.sh"))
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

type codexConfig struct {
	ApprovalPolicy     string `toml:"approval_policy"`
	SandboxMode        string `toml:"sandbox_mode"`
	ProjectDocMaxBytes int64  `toml:"project_doc_max_bytes"`
	AllowLoginShell    bool   `toml:"allow_login_shell"`
}

func analyseCodexConfigurations(agentHome, source string, output io.Writer) ([]*configurationChange, error) {
	var changes []*configurationChange
	rulesChange, err := analyseCodexRulesConfiguration(agentHome, source, output)
	if err != nil {
		return nil, err
	}
	if rulesChange != nil {
		changes = append(changes, rulesChange)
	}
	hooksChange, err := analyseCodexHooksConfiguration(agentHome, output)
	if err != nil {
		return nil, err
	}
	if hooksChange != nil {
		changes = append(changes, hooksChange)
	}
	return changes, nil
}

func analyseCodexRulesConfiguration(agentHome, source string, output io.Writer) (*configurationChange, error) {
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

func analyseCodexHooksConfiguration(agentHome string, output io.Writer) (*configurationChange, error) {
	path := filepath.Join(agentHome, "hooks.json")
	original, mode, exists, err := readOptionalRegularFile(path)
	if err != nil {
		return nil, err
	}
	root := map[string]any{}
	if exists {
		if err := json.Unmarshal(original, &root); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		if root == nil {
			return nil, fmt.Errorf("parsing %s: top level must be a JSON object", path)
		}
	}
	changed, err := normalizeCodexToolGuard(root)
	if err != nil {
		return nil, fmt.Errorf("analysing %s: %w", path, err)
	}
	if !changed {
		fmt.Fprintln(output, "Configuration: Codex hooks already contain the SDLC tool guard.")
		return nil, nil
	}
	candidate, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding proposed %s: %w", path, err)
	}
	candidate = append(candidate, '\n')
	before := "hooks.json absent"
	if exists {
		before = "existing hooks preserved; JSON spacing and key order may be normalized"
	}
	return &configurationChange{
		path: path, beforeLabel: before, afterLabel: "hooks.PreToolUse adds or updates the SDLC tool guard",
		contents: candidate, mode: mode,
	}, nil
}

func normalizeCodexToolGuard(root map[string]any) (bool, error) {
	hooks, err := objectField(root, "hooks")
	if err != nil {
		return false, err
	}
	value, exists := hooks["PreToolUse"]
	var entries []any
	if exists {
		var ok bool
		entries, ok = value.([]any)
		if !ok {
			return false, errors.New("hooks.PreToolUse must be a JSON array")
		}
	}
	changed := false
	found := false
	for _, item := range entries {
		group, ok := item.(map[string]any)
		if !ok {
			return false, errors.New("hooks.PreToolUse entries must be JSON objects")
		}
		handlers, ok := group["hooks"].([]any)
		if !ok {
			continue
		}
		for _, handlerValue := range handlers {
			handler, ok := handlerValue.(map[string]any)
			if !ok || handler["command"] != toolGuardCommand {
				continue
			}
			found = true
			if group["matcher"] != "*" {
				group["matcher"] = "*"
				changed = true
			}
			if handler["type"] != "command" {
				handler["type"] = "command"
				changed = true
			}
			if handler["timeout"] != float64(5) {
				handler["timeout"] = 5
				changed = true
			}
		}
	}
	if !found {
		entries = append(entries, map[string]any{
			"matcher": "*",
			"hooks": []any{map[string]any{
				"type": "command", "command": toolGuardCommand, "timeout": 5,
				"statusMessage": "Checking SDLC tool policy",
			}},
		})
		changed = true
	}
	hooks["PreToolUse"] = entries
	root["hooks"] = hooks
	return changed, nil
}

func readOptionalRegularFile(path string) ([]byte, os.FileMode, bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0o600, false, nil
	}
	if err != nil {
		return nil, 0, false, fmt.Errorf("inspecting %s: %w", path, err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, 0, false, fmt.Errorf("configuration path %s must be a regular non-symlink file", path)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, false, fmt.Errorf("reading %s: %w", path, err)
	}
	return contents, info.Mode().Perm(), true, nil
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

func offerConfigurationChanges(changes []*configurationChange, input io.Reader, output io.Writer) error {
	if len(changes) == 0 {
		fmt.Fprintln(output, "Configuration: no automatic change is available for this target.")
		return nil
	}
	for _, change := range changes {
		printConfigurationChange(output, change)
	}
	fmt.Fprint(output, "Apply these configuration changes? Type yes to continue: ")
	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading configuration confirmation: %w", err)
		}
		fmt.Fprintln(output, "Configuration unchanged.")
		return nil
	}
	if !confirmationAccepted(scanner.Text()) {
		fmt.Fprintln(output, "Configuration unchanged.")
		return nil
	}
	for _, change := range changes {
		if err := applyConfigurationChange(change, output); err != nil {
			return err
		}
	}
	return nil
}

func printConfigurationChange(output io.Writer, change *configurationChange) {
	fmt.Fprintf(output, "Proposed configuration change: %s\n", change.path)
	fmt.Fprintf(output, "- before: %s\n+ after: %s\n", change.beforeLabel, change.afterLabel)
}

func applyConfigurationChange(change *configurationChange, output io.Writer) error {
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
