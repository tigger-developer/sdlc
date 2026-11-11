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
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/tigger-developer/sdlc/internal/harness"
	"github.com/tigger-developer/sdlc/internal/skillinventory"
)

const (
	agentAuto    = "auto"
	agentClaude  = "claude"
	agentCodex   = "codex"
	agentCopilot = "copilot"
	agentHermes  = "hermes"
	agentGoose   = "goose"
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
	{name: agentGoose, skills: true},
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
	"audit-definition",
	"audit-implementation",
	"audit-project-foundation",
	"audit-test-code",
	"audit-test-definitions",
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
	filepath.Join("bin", "sdlc-preview"),
	filepath.Join("presets", "sdlc-standards"),
	filepath.Join("templates", "project-init"),
}

var retiredGlobalSkillPaths = []string{
	"audit-acs",
	"audit-tests",
	"design-solution",
	"draft-bug-fix",
	"draft-design-issue",
	"draft-issue",
}

var deployedCommandNames = []string{
	"sdlc-validate",
	"sdlc-init",
	"sdlc-merge-legacy-acs",
}

var deployedInternalCommandNames = []string{
	"sdlc-harness",
	"sdlc-guard-shell",
}

const (
	canonicalInstallRoot  = "/opt/sdlc"
	codexPythonRulesStart = "# BEGIN SDLC MANAGED PYTHON RULES"
	codexPythonRulesEnd   = "# END SDLC MANAGED PYTHON RULES"
	toolGuardCommand      = "bash /opt/sdlc/hooks/agent-command-guard.sh"
)

type Options struct {
	Agent       string
	AgentHome   string
	Source      string
	Apply       bool
	Configure   bool
	Release     string
	InstallRoot string
	Input       io.Reader
	Output      io.Writer
}

type configurationChange struct {
	path        string
	beforeLabel string
	afterLabel  string
	contents    []byte
	mode        os.FileMode
	conflict    bool
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

type managedLink struct {
	target            string
	destination       string
	destinationRoot   string
	ownership         linkOwnership
	needsSync         bool
	destinationExists bool
}

type installationPlan struct {
	inventories           []skillinventory.Plan
	syncs                 []managedSync
	links                 []managedLink
	retirements           []managedRetirement
	managedConfigurations []*configurationChange
	configurations        []*configurationChange
}

func Run(options Options) error {
	options = withDefaultIO(options)
	installRoot := options.InstallRoot
	if installRoot == "" {
		installRoot = canonicalInstallRoot
	}
	source, err := validateSource(options.Source)
	if err != nil {
		return err
	}
	agent, agentHome, err := resolveTarget(options.Agent, options.AgentHome, source)
	if err != nil {
		return err
	}
	if options.Apply || options.Configure {
		if err := ensureInstallRoot(installRoot, options.Input, options.Output); err != nil {
			return err
		}
	}
	plan, err := planInstallation(agent, source, agentHome, installRoot, options.Release)
	if err != nil {
		return err
	}

	printHeading(options.Output, agent, source, agentHome, options.Apply, options.Configure)
	printInstallationPlan(options.Output, plan, options.Apply)
	changes, err := analyseConfiguration(agent, agentHome, source, installRoot, options.Output)
	if err != nil {
		return err
	}
	if options.Apply {
		if err := applyInstallation(plan, options.Output); err != nil {
			return err
		}
		if err := ensureSharedDirectoryModes(installRoot); err != nil {
			return err
		}
		verified, verifyErr := planInstallation(agent, source, agentHome, installRoot, options.Release)
		if verifyErr != nil {
			return fmt.Errorf("verifying installation: %w", verifyErr)
		}
		if installationHasChanges(verified) {
			return errors.New("installation verification still reports changes")
		}
		if err := verifyCanonicalMain(installRoot); err != nil {
			return err
		}
		fmt.Fprintln(options.Output, "Installation: synchronized all planned copies.")
	}
	if options.Configure {
		if err := offerConfigurationChanges(changes, options.Input, options.Output); err != nil {
			return err
		}
		return ensureSharedDirectoryModes(installRoot)
	}
	if len(changes) != 0 {
		fmt.Fprintln(options.Output, "Configuration: recommendation only; re-run with --configure to review and apply it.")
	}
	return nil
}

func RunInteractive(sourcePath, userHome, release string, input io.Reader, output io.Writer) error {
	return runInteractiveAtRoot(sourcePath, userHome, canonicalInstallRoot, release, input, output)
}

// RunInteractiveAtRoot is for isolated installer verification; production uses RunInteractive.
func RunInteractiveAtRoot(sourcePath, userHome, installRoot, release string, input io.Reader, output io.Writer) error {
	return runInteractiveAtRoot(sourcePath, userHome, installRoot, release, input, output)
}

func runInteractiveAtRoot(sourcePath, userHome, installRoot, release string, input io.Reader, output io.Writer) error {
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
	reader := bufio.NewReader(input)
	if err := ensureInstallRoot(installRoot, reader, output); err != nil {
		return err
	}
	agents, err := selectInteractiveAgents(userHome, installRoot, reader, output)
	if err != nil {
		return err
	}
	plan := installationPlan{}
	shared, planErr := planSharedInstallation(source, userHome, installRoot, release)
	if planErr != nil {
		return planErr
	}
	mergeInstallationPlan(&plan, shared)
	for _, agent := range agents {
		agentHome := providerHome(userHome, agent)
		provider, planErr := planProviderInstallation(agent, source, agentHome, installRoot)
		if planErr != nil {
			return planErr
		}
		mergeInstallationPlan(&plan, provider)
	}
	configurations, err := planDetectedConfigurations(source, userHome, installRoot, agents)
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
		if err := verifyHarnessGuards(installRoot, agents, output); err != nil {
			return err
		}
		if err := verifyHarnessAdapters(installRoot, agents); err != nil {
			return err
		}
		printHarnessLocalValidation(output, agents)
		return nil
	}
	printInstallationPlan(output, plan, false)
	for _, change := range plan.configurations {
		printConfigurationChange(output, change)
		if change.conflict {
			accepted, confirmErr := confirm(reader, output, "Replace this unknown same-path conflict after moving the existing file to Trash? [yes/no]: ")
			if confirmErr != nil {
				return confirmErr
			}
			if !accepted {
				return fmt.Errorf("declined replacement leaves the provider not SDLC-ready: %s", change.path)
			}
		}
	}
	if err := applyInstallation(plan, output); err != nil {
		return err
	}
	for _, change := range plan.configurations {
		if err := applyConfigurationChange(change, output); err != nil {
			return err
		}
	}
	if err := ensureSharedDirectoryModes(installRoot); err != nil {
		return err
	}
	verified, verifyErr := planDetectedInstallation(source, userHome, installRoot, agents, release)
	if verifyErr != nil {
		return verifyErr
	}
	// Configuration recommendations are intentionally reviewed separately and
	// are not evidence that the file deployment failed verification.
	verified.configurations = nil
	verified.managedConfigurations = nil
	if installationHasChanges(verified) {
		return errors.New("deployment verification still reports changes")
	}
	if err := verifyCanonicalMain(installRoot); err != nil {
		return err
	}
	if err := verifyHarnessGuards(installRoot, agents, output); err != nil {
		return err
	}
	if err := verifyHarnessAdapters(installRoot, agents); err != nil {
		return err
	}
	printHarnessLocalValidation(output, agents)
	fmt.Fprintln(output, "All listed SDLC changes were installed.")
	return nil
}

func planDetectedInstallation(source, userHome, installRoot string, agents []string, release string) (installationPlan, error) {
	plan := installationPlan{}
	shared, err := planSharedInstallation(source, userHome, installRoot, release)
	if err != nil {
		return installationPlan{}, err
	}
	mergeInstallationPlan(&plan, shared)
	for _, agent := range agents {
		provider, err := planProviderInstallation(agent, source, providerHome(userHome, agent), installRoot)
		if err != nil {
			return installationPlan{}, err
		}
		mergeInstallationPlan(&plan, provider)
	}
	configurations, err := planDetectedConfigurations(source, userHome, installRoot, agents)
	if err != nil {
		return installationPlan{}, err
	}
	plan.configurations = configurations
	return plan, nil
}

func planDetectedConfigurations(source, userHome, installRoot string, agents []string) ([]*configurationChange, error) {
	var changes []*configurationChange
	for _, agent := range agents {
		agentChanges, err := analyseConfiguration(agent, providerHome(userHome, agent), source, installRoot, io.Discard)
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
		path := providerHome(userHome, provider.name)
		info, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspecting %s agent home %q: %w", provider.name, path, err)
		}
		if info.IsDir() {
			detected = append(detected, provider.name)
		}
	}
	return detected, nil
}

func printHarnessLocalValidation(output io.Writer, agents []string) {
	for _, agent := range agents {
		if agent == agentGoose {
			fmt.Fprintln(output, "Harness goose: command-guard-projection=VALID; external-audit-adapter=NOT-IN-SCOPE; native-hook-execution=UNVERIFIED")
			continue
		}
		fmt.Fprintf(output, "Harness %s: command-guard-projection=VALID; external-audit-adapter-construction=VALID\n", agent)
	}
}

func verifyHarnessGuards(installRoot string, agents []string, output io.Writer) error {
	guard := filepath.Join(installRoot, "hooks", "agent-command-guard.sh")
	for _, agent := range agents {
		payload := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"rm obsolete"}}`
		if agent == agentCopilot {
			payload = `{"hookEventName":"preToolUse","toolName":"shell","toolArgs":{"command":"rm obsolete"}}`
		} else if agent == agentGoose {
			payload = `{"event":"PreToolUse","tool_name":"shell","tool_input":{"command":"rm obsolete"}}`
		} else if agent == agentHermes {
			payload = `{"hook_event_name":"pre_tool_call","tool_name":"terminal","tool_input":{"command":"rm obsolete"}}`
		}
		command := exec.Command(guard) // #nosec G204 -- guard is the exact managed installation path.
		command.Stdin = strings.NewReader(payload)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		err := command.Run()
		verified := false
		if agent == agentHermes {
			var response map[string]any
			verified = err == nil && json.Unmarshal(stdout.Bytes(), &response) == nil && response["decision"] == "block"
		} else {
			var exitError *exec.ExitError
			verified = errors.As(err, &exitError) && exitError.ExitCode() == 2 && strings.Contains(stderr.String(), "Blocked by agent-command-guard")
		}
		if !verified {
			fmt.Fprintf(output, "Harness %s: command-guard-projection=INVALID; guard verification failed at %s\n", agent, guard)
			return fmt.Errorf("%s native command-guard projection could not be verified", agent)
		}
	}
	return nil
}

func verifyHarnessAdapters(installRoot string, agents []string) error {
	legacy := false
	for _, agent := range agents {
		legacy = legacy || agent != agentGoose
	}
	if !legacy {
		return nil
	}
	runner := filepath.Join(installRoot, "bin", "sdlc-harness")
	info, err := os.Stat(runner)
	if err != nil {
		return fmt.Errorf("verifying installed harness runner %q: %w", runner, err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("verifying installed harness runner %q: expected an executable regular file", runner)
	}
	for _, harnessName := range agents {
		if harnessName == agentGoose {
			continue // Goose support does not revive the abandoned external audit harness.
		}
		request := harness.Request{
			Harness: harnessName, Model: "readiness-model", Provider: "readiness-provider",
			Prompt: "readiness", Directory: "/project", ResultFile: "/result",
			SessionID: "00000000-0000-4000-8000-000000000000",
		}
		if _, err := harness.BuildStart(request); err != nil {
			return fmt.Errorf("verifying %s external-audit start adapter: %w", harnessName, err)
		}
		if _, err := harness.BuildResume(request); err != nil {
			return fmt.Errorf("verifying %s external-audit resume adapter: %w", harnessName, err)
		}
	}
	return nil
}

func installationHasChanges(plan installationPlan) bool {
	for _, inventory := range plan.inventories {
		if inventory.Changed || len(inventory.Retirements) != 0 {
			return true
		}
	}
	if len(plan.configurations) != 0 || len(plan.managedConfigurations) != 0 {
		return true
	}
	for _, retirement := range plan.retirements {
		if _, err := os.Lstat(retirement.path); err == nil {
			return true
		}
	}
	for _, sync := range plan.syncs {
		if sync.needsSync {
			return true
		}
	}
	for _, link := range plan.links {
		if link.needsSync {
			return true
		}
	}
	return false
}

func mergeInstallationPlan(plan *installationPlan, addition installationPlan) {
	plan.inventories = append(plan.inventories, addition.inventories...)
	plan.syncs = append(plan.syncs, addition.syncs...)
	plan.links = append(plan.links, addition.links...)
	plan.retirements = append(plan.retirements, addition.retirements...)
	plan.managedConfigurations = append(plan.managedConfigurations, addition.managedConfigurations...)
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
		return "", "", fmt.Errorf("unsupported agent %q; use auto, claude, codex, copilot, hermes, goose, or custom", agent)
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
	case agentGoose:
		return agentGoose
	default:
		return ""
	}
}

func defaultAgentHome(agent string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving user home: %w", err)
	}
	return providerHome(home, agent), nil
}

func providerHome(userHome, agent string) string {
	if agent == agentGoose {
		return filepath.Join(userHome, ".config", "goose")
	}
	return filepath.Join(userHome, "."+agent)
}

func commonAgentHome(agent, agentHome string) string {
	parent := filepath.Dir(agentHome)
	if agent == agentGoose && filepath.Base(parent) == ".config" {
		parent = filepath.Dir(parent)
	}
	return filepath.Join(parent, ".agents")
}

func planInstallation(agent, source, agentHome, installRoot, release string) (installationPlan, error) {
	userHome := filepath.Dir(commonAgentHome(agent, agentHome))
	plan := installationPlan{}
	shared, err := planSharedInstallation(source, userHome, installRoot, release)
	if err != nil {
		return installationPlan{}, err
	}
	mergeInstallationPlan(&plan, shared)
	provider, err := planProviderInstallation(agent, source, agentHome, installRoot)
	if err != nil {
		return installationPlan{}, err
	}
	mergeInstallationPlan(&plan, provider)
	return plan, nil
}

func planSharedInstallation(source, userHome, installRoot, release string) (installationPlan, error) {
	return planSharedInstallationMode(source, userHome, installRoot, release, false)
}

func planSharedInstallationMode(source, userHome, installRoot, release string, bare bool) (installationPlan, error) {
	liveSDLC := filepath.Join(installRoot, "docs")
	localBin := filepath.Join(userHome, ".local", "bin")
	if sameLexicalPath(source, installRoot) {
		return installationPlan{}, fmt.Errorf("source %q is the live SDLC directory; use a separate staging clone", source)
	}
	plan := installationPlan{}
	for _, mapping := range []struct {
		source, destination string
		optional            bool
	}{
		{source: filepath.Join(source, "src"), destination: liveSDLC},
		{source: filepath.Join(source, "commands"), destination: filepath.Join(installRoot, "commands"), optional: true},
		{source: filepath.Join(source, "skills"), destination: filepath.Join(installRoot, "skills")},
		{source: filepath.Join(source, "hooks"), destination: filepath.Join(installRoot, "hooks")},
	} {
		var files []managedSync
		var err error
		if filepath.Base(mapping.source) == "skills" {
			files, err = planSkillDirectoryMapping(mapping.source, mapping.destination, mapping.optional)
		} else {
			files, err = planDirectoryMapping(mapping.source, mapping.destination, mapping.optional)
		}
		if err != nil {
			return installationPlan{}, err
		}
		plan.syncs = append(plan.syncs, files...)
	}
	currentSkills, err := sourceDirectoryNames(filepath.Join(source, "skills"))
	if err != nil {
		return installationPlan{}, err
	}
	globalSkills := filepath.Join(userHome, ".agents", "skills")
	if err := plan.addSkillInventory(source, filepath.Join(installRoot, "skills"), "", currentSkills); err != nil {
		return installationPlan{}, err
	}
	if !bare {
		if err := plan.addSkillInventory(source, globalSkills, filepath.Join(installRoot, "skills"), currentSkills); err != nil {
			return installationPlan{}, err
		}
	}
	for _, skill := range currentSkills {
		if bare {
			continue
		}
		if isAuditSkill(skill) {
			continue
		}
		link, linkErr := planManagedLink(filepath.Join(installRoot, "skills", skill), filepath.Join(globalSkills, skill), globalSkills, skillLinkOwnership(source, userHome, skill, filepath.Join(source, "skills", skill)))
		if linkErr != nil {
			return installationPlan{}, linkErr
		}
		plan.links = append(plan.links, link)
	}
	for _, name := range deployedCommandNames {
		sourcePath := filepath.Join(source, "bin", name)
		if info, err := os.Stat(sourcePath); err != nil {
			return installationPlan{}, fmt.Errorf("inspecting built command %q: %w", sourcePath, err)
		} else if !info.Mode().IsRegular() {
			return installationPlan{}, fmt.Errorf("built command %q is not a regular file", sourcePath)
		}
		deployedPath := filepath.Join(installRoot, "bin", name)
		sync, err := planFileSync(sourcePath, deployedPath, filepath.Join(installRoot, "bin"))
		if err != nil {
			return installationPlan{}, err
		}
		plan.syncs = append(plan.syncs, sync)
		if bare {
			continue
		}
		link, err := planManagedLink(deployedPath, filepath.Join(localBin, name), localBin, linkOwnership{legacyTargets: []string{filepath.Join(userHome, ".agents", "sdlc", "bin", name), sourcePath}})
		if err != nil {
			return installationPlan{}, err
		}
		plan.links = append(plan.links, link)
	}
	for _, name := range deployedInternalCommandNames {
		sourcePath := filepath.Join(source, "bin", name)
		if info, err := os.Stat(sourcePath); err != nil {
			return installationPlan{}, fmt.Errorf("inspecting built internal command %q: %w", sourcePath, err)
		} else if !info.Mode().IsRegular() {
			return installationPlan{}, fmt.Errorf("built internal command %q is not a regular file", sourcePath)
		}
		deployedPath := filepath.Join(installRoot, "bin", name)
		sync, err := planFileSync(sourcePath, deployedPath, filepath.Join(installRoot, "bin"))
		if err != nil {
			return installationPlan{}, err
		}
		plan.syncs = append(plan.syncs, sync)
	}
	for _, retirement := range []struct {
		sourceRoot, destinationRoot string
		files                       []string
	}{
		{filepath.Join(source, "src"), liveSDLC, retiredSharedFiles},
		{filepath.Join(source, "commands"), filepath.Join(installRoot, "commands"), retiredCommandFiles},
		{filepath.Join(source, "skills"), filepath.Join(installRoot, "skills"), retiredSkillFiles},
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
	retirements, retirementErr := planExactRetirements(filepath.Join(installRoot, "skills"), retiredGlobalSkillPaths)
	if retirementErr != nil {
		return installationPlan{}, retirementErr
	}
	plan.retirements = append(plan.retirements, retirements...)
	if !bare {
		if err := rejectUnverifiedRetiredSkills(globalSkills, retiredGlobalSkillPaths); err != nil {
			return installationPlan{}, err
		}
	}
	auditRetirements, auditErr := planAuditSkillRetirements(filepath.Join(installRoot, "skills"))
	if auditErr != nil {
		return installationPlan{}, auditErr
	}
	plan.retirements = append(plan.retirements, auditRetirements...)
	if !bare {
		auditLink, err := planManagedLink(filepath.Join(installRoot, "bin", "sdlc-harness"), filepath.Join(localBin, "sdlc-audit"), localBin, linkOwnership{legacyTargets: []string{filepath.Join(userHome, ".agents", "sdlc", "bin", "sdlc-harness")}})
		if err != nil {
			return installationPlan{}, err
		}
		plan.links = append(plan.links, auditLink)
		commandRetirements, err := planExactRetirements(localBin, []string{
			"sdlc-install",
			"sdlc-preview",
			"sdlc-project-init",
			"sdlc-project-update",
		})
		if err != nil {
			return installationPlan{}, err
		}
		plan.retirements = append(plan.retirements, commandRetirements...)
	}
	for _, root := range []string{localBin, filepath.Join(installRoot, "bin")} {
		if bare && root == localBin {
			continue
		}
		retired, err := planRetiredCommandBackups(root)
		if err != nil {
			return installationPlan{}, err
		}
		plan.retirements = append(plan.retirements, retired...)
	}
	if bare {
		pluginChanges, err := analyseGooseConfiguration(installRoot)
		if err != nil {
			return installationPlan{}, err
		}
		for _, change := range pluginChanges {
			if change.conflict {
				return installationPlan{}, fmt.Errorf("bare installation cannot replace conflicting Goose plugin file %q; review it before retrying", change.path)
			}
		}
		plan.managedConfigurations = append(plan.managedConfigurations, pluginChanges...)
	}
	releaseChange, err := planGlobalReleaseConfiguration(installRoot, userHome, release)
	if err != nil {
		return installationPlan{}, err
	}
	if releaseChange != nil {
		plan.managedConfigurations = append(plan.managedConfigurations, releaseChange)
	}
	return plan, nil
}

func verifyCanonicalMain(installRoot string) error {
	mainPath := filepath.Join(installRoot, "docs", "MAIN.md")
	info, err := os.Stat(mainPath)
	if err != nil {
		return fmt.Errorf("verifying canonical SDLC entry point %q: %w", mainPath, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("verifying canonical SDLC entry point %q: expected a regular file", mainPath)
	}
	return nil
}

func planProviderInstallation(agent, source, agentHome, installRoot string) (installationPlan, error) {
	plan := installationPlan{}
	if agent == agentCustom {
		return plan, nil
	}
	provider, found := providerDefinitionFor(agent)
	if !found {
		return installationPlan{}, fmt.Errorf("planning unsupported agent %q", agent)
	}
	if agent == agentGoose {
		userHome := filepath.Dir(filepath.Dir(agentHome))
		plugins := filepath.Join(userHome, ".agents", "plugins")
		link, err := planManagedLink(filepath.Join(installRoot, "plugins", "sdlc-command-guard-goose"), filepath.Join(plugins, "sdlc-command-guard-goose"), plugins, linkOwnership{kind: "goose-plugin"})
		if err != nil {
			return installationPlan{}, err
		}
		plan.links = append(plan.links, link)
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
	if provider.skills {
		currentSkills, planErr := sourceDirectoryNames(filepath.Join(source, "skills"))
		if planErr != nil {
			return installationPlan{}, planErr
		}
		canonicalSkills := filepath.Join(installRoot, "skills")
		providerSkills := filepath.Join(agentHome, "skills")
		if err := plan.addSkillInventory(source, providerSkills, canonicalSkills, currentSkills); err != nil {
			return installationPlan{}, err
		}
		for _, skill := range currentSkills {
			if isAuditSkill(skill) {
				continue
			}
			userHome := filepath.Dir(commonAgentHome(agent, agentHome))
			owner := skillLinkOwnership(source, userHome, skill, filepath.Join(source, "skills", skill))
			owner.legacyTargets = append(owner.legacyTargets, filepath.Join(userHome, ".agents", "skills", skill))
			link, linkErr := planManagedLink(filepath.Join(canonicalSkills, skill), filepath.Join(providerSkills, skill), providerSkills, owner)
			if linkErr != nil {
				return installationPlan{}, linkErr
			}
			plan.links = append(plan.links, link)
		}
		paths := append([]string{}, retiredProviderArtifacts...)
		for _, current := range currentSkills {
			if isAuditSkill(current) {
				paths = append(paths, current)
			}
		}
		for _, relative := range retiredSkillFiles {
			paths = append(paths, filepath.Dir(relative))
		}
		for _, current := range currentSkills {
			for index, path := range paths {
				if path == current {
					paths = append(paths[:index], paths[index+1:]...)
					break
				}
			}
		}
		if err := rejectUnverifiedRetiredSkills(providerSkills, paths); err != nil {
			return installationPlan{}, err
		}
	}
	if agent != agentCodex {
		if provider.commandPath != "" {
			commandRetirements, commandErr := planExactRetirements(filepath.Join(agentHome, provider.commandPath), retiredCommandFiles)
			if commandErr != nil {
				return installationPlan{}, commandErr
			}
			plan.retirements = append(plan.retirements, commandRetirements...)
		}
	}
	return plan, nil
}

func isAuditSkill(name string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(name)), "audit-")
}

func planSkillDirectoryMapping(sourceRoot, destinationRoot string, optional bool) ([]managedSync, error) {
	if optional {
		if _, err := os.Stat(sourceRoot); errors.Is(err, os.ErrNotExist) {
			return nil, nil
		} else if err != nil {
			return nil, fmt.Errorf("inspecting optional install source %q: %w", sourceRoot, err)
		}
	}
	names, err := sourceDirectoryNames(sourceRoot)
	if err != nil {
		return nil, err
	}
	var syncs []managedSync
	for _, name := range names {
		if isAuditSkill(name) {
			continue
		}
		files, err := planDirectoryFiles(filepath.Join(sourceRoot, name), filepath.Join(destinationRoot, name))
		if err != nil {
			return nil, err
		}
		syncs = append(syncs, files...)
	}
	return syncs, nil
}

func planAuditSkillRetirements(root string) ([]managedRetirement, error) {
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading managed skill directory %q: %w", root, err)
	}
	var retirements []managedRetirement
	for _, entry := range entries {
		// Backups are deliberately retained beside retired artefacts; they are
		// not live skills and must not make verification non-idempotent.
		if entry.IsDir() && !strings.Contains(entry.Name(), ".") && isAuditSkill(entry.Name()) && entry.Name() != "audit-tests" {
			retirements = append(retirements, managedRetirement{path: filepath.Join(root, entry.Name())})
		}
	}
	return retirements, nil
}

func sourceDirectoryNames(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("reading managed source directories %q: %w", root, err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		contents, err := directoryContents(path)
		if err != nil {
			return nil, fmt.Errorf("inspecting source skill %q: %w", path, err)
		}
		if len(contents.files) == 0 {
			continue
		}
		if _, present := contents.files["SKILL.md"]; !present {
			return nil, fmt.Errorf("source skill %q has files but no SKILL.md", path)
		}
		names = append(names, entry.Name())
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
	for _, inventory := range plan.inventories {
		for _, retirement := range inventory.Retirements {
			fmt.Fprintf(output, "Installation: trash inventoried retired skill %s\n", retirement.Path)
		}
		if inventory.Changed {
			fmt.Fprintf(output, "Installation: update skill inventory %s\n", inventory.Path)
		}
	}
	for _, sync := range plan.syncs {
		if !sync.needsSync {
			if os.Getenv("VERBOSE") == "1" {
				fmt.Fprintf(output, "Installation: current %s <- %s\n", sync.destination, sync.source)
			}
			continue
		}
		verb := "would install missing file"
		if sync.backupExisting {
			verb = "would trash and replace differing file"
		} else if sync.destinationExists {
			verb = "would trash and replace differing file"
		}
		if apply {
			verb = "will install missing file"
			if sync.backupExisting {
				verb = "will trash and replace differing file"
			} else if sync.destinationExists {
				verb = "will trash and replace differing file"
			}
		}
		fmt.Fprintf(output, "Installation: %s %s <- %s\n", verb, sync.destination, sync.source)
	}
	for _, retirement := range plan.retirements {
		verb := "would trash legacy SDLC artefact"
		if apply {
			verb = "will trash legacy SDLC artefact"
		}
		fmt.Fprintf(output, "Installation: %s %s\n", verb, retirement.path)
	}
	for _, link := range plan.links {
		if !link.needsSync {
			if os.Getenv("VERBOSE") == "1" {
				fmt.Fprintf(output, "Installation: current %s -> %s\n", link.destination, link.target)
			}
			continue
		}
		verb := "would install command link"
		if link.destinationExists {
			verb = "would trash and replace command link"
		}
		if apply {
			verb = "will install command link"
			if link.destinationExists {
				verb = "will trash and replace command link"
			}
		}
		fmt.Fprintf(output, "Installation: %s %s -> %s\n", verb, link.destination, link.target)
	}
	for _, change := range plan.managedConfigurations {
		verb := "would update managed configuration"
		if apply {
			verb = "will update managed configuration"
		}
		fmt.Fprintf(output, "Installation: %s %s\n", verb, change.path)
	}
}

func applyInstallation(plan installationPlan, output io.Writer) error {
	if installationHasChanges(plan) {
		if _, err := exec.LookPath("trash"); err != nil {
			return fmt.Errorf("installation requires the trash CLI on PATH: %w", err)
		}
	}
	for _, sync := range plan.syncs {
		if !sync.needsSync {
			continue
		}
		if err := synchronizeFile(sync, output); err != nil {
			return err
		}
	}
	for _, retirement := range plan.retirements {
		if err := trashArtifact(output, retirement.path); err != nil {
			return err
		}
		fmt.Fprintf(output, "Installation retired: %s\n", retirement.path)
	}
	for _, change := range plan.managedConfigurations {
		if err := applyConfigurationChange(change, output); err != nil {
			return err
		}
	}
	for _, link := range plan.links {
		if !link.needsSync {
			continue
		}
		if err := synchronizeLink(link, output); err != nil {
			return err
		}
	}
	for _, inventory := range plan.inventories {
		for _, retirement := range inventory.Retirements {
			if err := retirement.Validate(); err != nil {
				return err
			}
			if _, err := os.Lstat(retirement.Path); os.IsNotExist(err) {
				continue
			} else if err != nil {
				return err
			}
			if err := trashArtifact(output, retirement.Path); err != nil {
				return err
			}
		}
	}
	for _, inventory := range plan.inventories {
		if inventory.Changed {
			if err := writeFileAtomic(inventory.Path, inventory.Contents, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func (plan *installationPlan) addSkillInventory(source, destination, canonical string, names []string) error {
	var selected []string
	for _, name := range names {
		if !isAuditSkill(name) {
			selected = append(selected, name)
		}
	}
	inventory, err := skillinventory.Build(filepath.Join(source, "skills"), destination, canonical, selected, "sdlc")
	if err != nil {
		return err
	}
	plan.inventories = append(plan.inventories, inventory)
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

func planManagedLink(target, destination, destinationRoot string, ownership linkOwnership) (managedLink, error) {
	link := managedLink{target: target, destination: destination, destinationRoot: destinationRoot, ownership: ownership}
	obstructed, err := destinationParentObstructed(destination, destinationRoot)
	if err != nil {
		return managedLink{}, err
	}
	if obstructed {
		return managedLink{}, fmt.Errorf("cannot install SDLC link %q: a destination parent is not a regular directory", destination)
	}
	info, err := os.Lstat(destination)
	if errors.Is(err, os.ErrNotExist) {
		link.needsSync = true
		return link, nil
	}
	if err != nil {
		return managedLink{}, fmt.Errorf("inspecting command link %q: %w", destination, err)
	}
	link.destinationExists = true
	if info.Mode()&os.ModeSymlink == 0 {
		if err := verifyOwnedLinkReplacement(link, info); err != nil {
			return managedLink{}, err
		}
		link.needsSync = true
		return link, nil
	}
	current, err := os.Readlink(destination)
	if err != nil {
		return managedLink{}, fmt.Errorf("reading command link %q: %w", destination, err)
	}
	if !filepath.IsAbs(current) {
		current = filepath.Join(filepath.Dir(destination), current)
	}
	if !sameLexicalPath(current, target) {
		if err := verifyOwnedLinkReplacement(link, info); err != nil {
			return managedLink{}, err
		}
	}
	link.needsSync = !sameLexicalPath(current, target)
	return link, nil
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

func synchronizeFile(sync managedSync, output io.Writer) error {
	contents, err := os.ReadFile(sync.source)
	if err != nil {
		return fmt.Errorf("reading install source %q: %w", sync.source, err)
	}
	info, err := os.Stat(sync.source)
	if err != nil {
		return fmt.Errorf("inspecting install source %q: %w", sync.source, err)
	}
	if err := ensureRegularDirectory(filepath.Dir(sync.destination), output); err != nil {
		return err
	}
	if _, err := os.Lstat(sync.destination); err == nil {
		if err := trashArtifact(output, sync.destination); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspecting live file %q: %w", sync.destination, err)
	}
	if err := writeFileAtomic(sync.destination, contents, info.Mode().Perm()); err != nil {
		return err
	}
	fmt.Fprintf(output, "Installation updated: %s\n", sync.destination)
	return nil
}

func synchronizeLink(link managedLink, output io.Writer) error {
	obstructed, err := destinationParentObstructed(link.destination, link.destinationRoot)
	if err != nil {
		return err
	}
	if obstructed {
		return fmt.Errorf("cannot install SDLC link %q: a destination parent is not a regular directory", link.destination)
	}
	if err := ensureRegularDirectory(filepath.Dir(link.destination), output); err != nil {
		return err
	}
	if info, err := os.Lstat(link.destination); err == nil {
		if err := verifyOwnedLinkReplacement(link, info); err != nil {
			return err
		}
		if err := trashArtifact(output, link.destination); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspecting command link %q: %w", link.destination, err)
	}
	if err := os.Symlink(link.target, link.destination); err != nil {
		return fmt.Errorf("installing command link %q: %w", link.destination, err)
	}
	fmt.Fprintf(output, "Installation linked: %s -> %s\n", link.destination, link.target)
	return nil
}

func ensureRegularDirectory(directory string, output io.Writer) error {
	mode := os.FileMode(0o700)
	if directory == canonicalInstallRoot || strings.HasPrefix(directory, canonicalInstallRoot+string(os.PathSeparator)) {
		mode = 0o755
	}
	info, err := os.Lstat(directory)
	if err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		parent := filepath.Dir(directory)
		if parent != directory {
			if err := ensureRegularDirectory(parent, output); err != nil {
				return err
			}
		}
		if err := os.Mkdir(directory, mode); err != nil {
			return err
		}
		return os.Chmod(directory, mode)
	}
	if err != nil {
		return fmt.Errorf("inspecting installation directory %q: %w", directory, err)
	}
	if err := trashArtifact(output, directory); err != nil {
		return err
	}
	if err := os.Mkdir(directory, mode); err != nil {
		return err
	}
	return os.Chmod(directory, mode)
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

func analyseConfiguration(agent, agentHome, source, installRoot string, output io.Writer) ([]*configurationChange, error) {
	switch agent {
	case agentClaude:
		change, err := analyseClaudeConfiguration(agentHome, output)
		return configurationChanges(change), err
	case agentCodex:
		return analyseCodexConfigurations(agentHome, source, output)
	case agentGoose:
		return analyseGooseConfiguration(installRoot)
	case agentHermes:
		change, err := analyseHermesConfiguration(agentHome, source, output)
		return configurationChanges(change), err
	case agentCopilot:
		change, err := analyseCopilotConfiguration(agentHome, output)
		return configurationChanges(change), err
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
	settings, _, mode, err := readClaudeSettings(path)
	if err != nil {
		return nil, err
	}
	changed, err := normalizeClaudeToolGuard(settings)
	if err != nil {
		return nil, fmt.Errorf("analysing %s: %w", path, err)
	}
	if !changed {
		fmt.Fprintln(output, "Configuration: Claude settings already contain the SDLC tool guard.")
		return nil, nil
	}
	if settingsSymlink {
		return nil, fmt.Errorf("Claude configuration %s is a symlink; update its target explicitly", path)
	}
	candidate, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding proposed %s: %w", path, err)
	}
	candidate = append(candidate, '\n')
	return &configurationChange{
		path: path, beforeLabel: "existing configuration; unrelated values preserved",
		afterLabel: "PreToolUse invokes the canonical SDLC command guard", contents: candidate, mode: mode,
	}, nil
}

func normalizeClaudeToolGuard(root map[string]any) (bool, error) {
	return normalizeJSONToolGuard(root, false)
}

func analyseCopilotConfiguration(agentHome string, output io.Writer) (*configurationChange, error) {
	path := filepath.Join(agentHome, "hooks", "sdlc-tool-guard.json")
	original, mode, exists, err := readOptionalRegularFile(path)
	if err != nil {
		return nil, err
	}
	candidate, err := json.MarshalIndent(map[string]any{
		"version": 1,
		"hooks": map[string]any{"preToolUse": []any{map[string]any{
			"type": "command", "command": toolGuardCommand, "timeoutSec": 5,
		}}},
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	candidate = append(candidate, '\n')
	if exists {
		if migrated, ok := migrateCopilotGuard(original); ok {
			candidate = migrated
		}
	}
	if exists && bytes.Equal(original, candidate) {
		fmt.Fprintln(output, "Configuration: Copilot hooks already contain the SDLC tool guard.")
		return nil, nil
	}
	before := "managed hook file absent"
	conflict := false
	if exists {
		if _, managed := migrateCopilotGuard(original); managed {
			before = "managed hook file differs and will be backed up"
		} else {
			before = "unknown same-path conflict; unrelated content is not presumed SDLC-owned"
			conflict = true
		}
	}
	return &configurationChange{path: path, beforeLabel: before, afterLabel: "preToolUse invokes the canonical SDLC command guard", contents: candidate, mode: mode, conflict: conflict}, nil
}

// Migrate only recognized guards, retaining other hooks and configuration fields.
func migrateCopilotGuard(contents []byte) ([]byte, bool) {
	var root map[string]any
	if json.Unmarshal(contents, &root) != nil || root["version"] != float64(1) {
		return nil, false
	}
	hooks, ok := root["hooks"].(map[string]any)
	if !ok {
		return nil, false
	}
	entries, ok := hooks["preToolUse"].([]any)
	if !ok {
		return nil, false
	}
	var retained []any
	found := false
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		command, _ := entry["command"].(string)
		if ok && isManagedGuardCommand(command) {
			found = true
		} else {
			retained = append(retained, raw)
		}
	}
	if !found {
		return nil, false
	}
	hooks["preToolUse"] = append(retained, map[string]any{"type": "command", "command": toolGuardCommand, "timeoutSec": 5})
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, false
	}
	return append(data, '\n'), true
}

func onlyKeys(value map[string]any, allowed ...string) bool {
	if len(value) != len(allowed) {
		return false
	}
	for _, key := range allowed {
		if _, exists := value[key]; !exists {
			return false
		}
	}
	return true
}

func configurationChanges(change *configurationChange) []*configurationChange {
	if change == nil {
		return nil
	}
	return []*configurationChange{change}
}

func isManagedGuardCommand(command string) bool {
	fields := strings.Fields(strings.TrimSpace(command))
	if len(fields) != 2 || filepath.Base(strings.Trim(fields[0], "\"'")) != "bash" {
		return false
	}
	path := strings.Trim(fields[1], "\"'")
	return path == "/opt/sdlc/hooks/agent-command-guard.sh" ||
		path == "~/.agents/sdlc/hooks/agent-command-guard.sh" ||
		strings.HasSuffix(path, "/.agents/sdlc/hooks/agent-command-guard.sh") ||
		strings.HasSuffix(path, "/.hermes/sdlc/hooks/agent-command-guard.sh") ||
		strings.HasSuffix(path, "/.claude/sdlc/hooks/agent-command-guard.sh")
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
	return normalizeJSONToolGuard(root, true)
}

func normalizeJSONToolGuard(root map[string]any, statusMessage bool) (bool, error) {
	before, err := json.Marshal(root)
	if err != nil {
		return false, err
	}
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
	var normalized []any
	for _, item := range entries {
		group, ok := item.(map[string]any)
		if !ok {
			return false, errors.New("hooks.PreToolUse entries must be JSON objects")
		}
		handlers, ok := group["hooks"].([]any)
		if !ok {
			normalized = append(normalized, group)
			continue
		}
		var retained []any
		removed := false
		for _, handlerValue := range handlers {
			handler, ok := handlerValue.(map[string]any)
			command, _ := handler["command"].(string)
			if ok && isManagedGuardCommand(command) {
				removed = true
			} else {
				retained = append(retained, handlerValue)
			}
		}
		if !removed {
			normalized = append(normalized, group)
		} else if len(retained) != 0 {
			group["hooks"] = retained
			normalized = append(normalized, group)
		}
	}
	// Isolate the global guard so unrelated handlers retain their original matcher.
	guard := map[string]any{"type": "command", "command": toolGuardCommand, "timeout": 5}
	if statusMessage {
		guard["statusMessage"] = "Checking SDLC tool policy"
	}
	normalized = append(normalized, map[string]any{
		"matcher": "*",
		"hooks":   []any{guard},
	})
	hooks["PreToolUse"] = normalized
	root["hooks"] = hooks
	after, err := json.Marshal(root)
	return !bytes.Equal(before, after), err
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
	reader := bufio.NewReader(input)
	for _, change := range changes {
		printConfigurationChange(output, change)
		if change.conflict {
			accepted, err := confirm(reader, output, "Replace this unknown same-path conflict after moving the existing file to Trash? [yes/no]: ")
			if err != nil {
				return err
			}
			if !accepted {
				return fmt.Errorf("declined replacement leaves the provider not SDLC-ready: %s", change.path)
			}
		}
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
	return applyConfigurationChangeWith(change, output, writeFileAtomic, trashArtifact)
}

func applyConfigurationChangeWith(change *configurationChange, output io.Writer, write func(string, []byte, os.FileMode) error, trash func(io.Writer, string) error) error {
	if _, err := os.Lstat(change.path); err == nil {
		if err := trash(output, change.path); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspecting configuration %q: %w", change.path, err)
	}
	if err := write(change.path, change.contents, change.mode); err != nil {
		return fmt.Errorf("writing configuration %q after preserving the previous file in Trash: %w", change.path, err)
	}
	fmt.Fprintf(output, "Configuration updated: %s\n", change.path)
	return nil
}

func writeFileAtomic(path string, contents []byte, mode os.FileMode) error {
	directory := filepath.Dir(path)
	if err := ensureRegularDirectory(directory, io.Discard); err != nil {
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
