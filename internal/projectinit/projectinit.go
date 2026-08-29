package projectinit

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	keyAgentHarness        = "SDLC_AGENT_HARNESS"
	keySpecProvider        = "SDLC_SPEC_PROVIDER"
	keySpecModel           = "SDLC_SPEC_MODEL"
	keyBuildProvider       = "SDLC_BUILD_PROVIDER"
	keyBuildModel          = "SDLC_BUILD_MODEL"
	keyAuditProvider       = "SDLC_AUDIT_PROVIDER"
	keyAuditModel          = "SDLC_AUDIT_MODEL"
	keyInfraEnabled        = "SDLC_INFRA_ENABLED"
	keyInfraOwner          = "SDLC_INFRA_OWNER"
	keyInfraContract       = "SDLC_INFRA_CONTRACT"
	keyTechnologies        = "SDLC_TECHNOLOGIES"
	legacyDeliveryProvider = "SDLC_DELIVERY_PROVIDER"
	legacyDeliveryModel    = "SDLC_DELIVERY_MODEL"
)

var managedKeys = []string{
	keyAgentHarness,
	keySpecProvider,
	keySpecModel,
	keyBuildProvider,
	keyBuildModel,
	keyAuditProvider,
	keyAuditModel,
	keyInfraEnabled,
	keyInfraOwner,
	keyInfraContract,
	keyTechnologies,
}

var universalStandards = []standard{
	{Path: "MAIN.md", Subject: "Universal engineering behaviour"},
	{Path: "ISSUES.md", Subject: "Specification and requirement quality"},
	{Path: "CODING.md", Subject: "Implementation and design"},
	{Path: "TESTING.md", Subject: "Testing and evidence"},
	{Path: "DOCUMENTATION.md", Subject: "Documentation"},
	{Path: "GIT.md", Subject: "Source control"},
}

type standard struct {
	Path    string
	Subject string
}

// Technology is one selectable engineering-standard document.
type Technology struct {
	Name  string
	Title string
	Path  string
}

// Options controls one project initialization run.
type Options struct {
	ProjectRoot    string
	SDLCRoot       string
	UserConfigPath string
	Harness        string
	SpecProvider   string
	SpecModel      string
	BuildProvider  string
	BuildModel     string
	AuditProvider  string
	AuditModel     string
	Technologies   []string
	InfraEnabled   *bool
	InfraOwner     string
	InfraContract  string
	SDLCRevision   string
	NoLaunch       bool
	Input          io.Reader
	Output         io.Writer
	ErrorOutput    io.Writer
	RunCommand     func(string, []string, string, io.Reader, io.Writer, io.Writer) error
}

type resolvedConfig struct {
	Harness       string
	SpecProvider  string
	SpecModel     string
	BuildProvider string
	BuildModel    string
	AuditProvider string
	AuditModel    string
	Technologies  []string
	InfraEnabled  *bool
	InfraOwner    string
	InfraContract string
	SDLCRevision  string
}

// Run renders the deterministic constitution baseline and, when it changes,
// invokes the selected agent harness for project-specific completion.
func Run(options Options) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving user home: %w", err)
	}
	useCanonicalReference := options.SDLCRoot == ""
	if options.SDLCRoot == "" {
		options.SDLCRoot = filepath.Join(home, ".agents", "sdlc")
	}
	if options.UserConfigPath == "" {
		options.UserConfigPath = userDefaultsPath(home)
	}
	options = defaultOptions(options)
	projectRoot, err := filepath.Abs(options.ProjectRoot)
	if err != nil {
		return fmt.Errorf("resolving project root: %w", err)
	}
	sdlcRoot, err := filepath.Abs(options.SDLCRoot)
	if err != nil {
		return fmt.Errorf("resolving SDLC root: %w", err)
	}

	reader := bufio.NewReader(options.Input)
	projectConfigPath := filepath.Join(projectRoot, ".env")
	userValues, err := readManagedEnv(options.UserConfigPath)
	if err != nil {
		return err
	}
	projectValues, err := readManagedEnv(projectConfigPath)
	if err != nil {
		return err
	}
	normalizeLegacyDelivery(userValues)
	legacyProjectConfig := normalizeLegacyDelivery(projectValues)
	config := resolveConfig(options, userValues, projectValues)
	configChanged := legacyProjectConfig

	if err := ensureSpecKit(projectRoot, &config, reader, options); err != nil {
		return err
	}
	if err := ensurePreset(projectRoot, sdlcRoot, options); err != nil {
		return err
	}
	technologies, err := DiscoverTechnologies(filepath.Join(sdlcRoot, "technologies"))
	if err != nil {
		return err
	}
	if len(config.Technologies) == 0 {
		selected, promptErr := promptTechnologies(reader, options.Output, technologies)
		if promptErr != nil {
			return promptErr
		}
		config.Technologies = selected
		projectValues[keyTechnologies] = strings.Join(selected, ",")
		configChanged = true
	}
	selected, err := selectTechnologies(technologies, config.Technologies)
	if err != nil {
		return err
	}
	if config.InfraEnabled == nil {
		enabled, promptErr := promptYesNo(reader, options.Output, "Does an external project own deployment or runtime infrastructure? [yes/no]: ")
		if promptErr != nil {
			return promptErr
		}
		config.InfraEnabled = &enabled
		projectValues[keyInfraEnabled] = fmt.Sprintf("%t", enabled)
		configChanged = true
	}
	if *config.InfraEnabled {
		if config.InfraOwner == "" {
			config.InfraOwner, err = promptRequired(reader, options.Output, "Infrastructure owner descriptor: ")
			if err != nil {
				return err
			}
			projectValues[keyInfraOwner] = config.InfraOwner
			configChanged = true
		}
		if config.InfraContract == "" {
			config.InfraContract, err = promptRequired(reader, options.Output, "Infrastructure integration-contract path: ")
			if err != nil {
				return err
			}
			projectValues[keyInfraContract] = config.InfraContract
			configChanged = true
		}
	}
	if config.Harness == "" {
		config.Harness, err = promptRequired(reader, options.Output, "Agent harness (codex, claude, or hermes): ")
		if err != nil {
			return err
		}
		projectValues[keyAgentHarness] = config.Harness
		configChanged = true
	}
	if err := validateHarness(config.Harness); err != nil {
		return err
	}
	if err := validateProviderModelPairs(config); err != nil {
		return err
	}
	for key, value := range projectSnapshot(config) {
		if _, exists := projectValues[key]; !exists {
			projectValues[key] = value
			configChanged = true
		}
	}

	referenceRoot := sdlcRoot
	if useCanonicalReference {
		referenceRoot = "~/.agents/sdlc"
	}
	rendered := renderConstitution(referenceRoot, selected, config)
	target := filepath.Join(projectRoot, ".specify", "templates", "overrides", "constitution-template.md")
	current, readErr := os.ReadFile(target)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return fmt.Errorf("reading constitution template %q: %w", target, readErr)
	}
	templateChanged := readErr != nil || !bytes.Equal(current, rendered)

	if configChanged {
		if err := writeManagedEnv(projectConfigPath, projectValues); err != nil {
			return err
		}
		fmt.Fprintf(options.Output, "Updated SDLC project selections: %s\n", projectConfigPath)
		ignorePath := filepath.Join(projectRoot, ".gitignore")
		ignoreChanged, err := ensureEnvIgnored(ignorePath)
		if err != nil {
			return err
		}
		if ignoreChanged {
			fmt.Fprintf(options.Output, "Added .env ignore rule: %s\n", ignorePath)
		}
	}
	if templateChanged {
		if err := writeAtomic(target, rendered, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(options.Output, "Updated constitution baseline: %s\n", target)
		if os.Getenv("VERBOSE") == "1" {
			printVariance(options.Output, current, rendered)
		}
	}
	if err := commitConstitutionBaseline(projectRoot, target, options); err != nil {
		return err
	}
	if !templateChanged && !configChanged {
		return nil
	}
	if options.NoLaunch {
		return nil
	}
	return launchConstitution(config, projectRoot, target, options)
}

func commitConstitutionBaseline(projectRoot, target string, options Options) error {
	relative, err := filepath.Rel(projectRoot, target)
	if err != nil {
		return fmt.Errorf("resolving constitution baseline path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("constitution baseline %q is outside project root %q", target, projectRoot)
	}
	relative = filepath.ToSlash(relative)

	var status bytes.Buffer
	statusArguments := []string{"status", "--porcelain=v1", "--untracked-files=all", "--", relative}
	if err := options.RunCommand("git", statusArguments, projectRoot, strings.NewReader(""), &status, options.ErrorOutput); err != nil {
		return fmt.Errorf("checking constitution baseline Git status: %w", err)
	}
	if strings.TrimSpace(status.String()) == "" {
		return nil
	}
	if err := options.RunCommand("git", []string{"add", "--", relative}, projectRoot, strings.NewReader(""), options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("staging constitution baseline: %w", err)
	}
	commitArguments := []string{"commit", "--only", "--quiet", "--message", "docs: update SDLC constitution baseline", "--", relative}
	if err := options.RunCommand("git", commitArguments, projectRoot, strings.NewReader(""), options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("committing constitution baseline: %w", err)
	}
	fmt.Fprintf(options.Output, "Committed constitution baseline: %s\n", relative)
	return nil
}

func defaultOptions(options Options) Options {
	if options.ProjectRoot == "" {
		options.ProjectRoot = "."
	}
	if options.Input == nil {
		options.Input = os.Stdin
	}
	if options.Output == nil {
		options.Output = os.Stdout
	}
	if options.ErrorOutput == nil {
		options.ErrorOutput = os.Stderr
	}
	if options.RunCommand == nil {
		options.RunCommand = runCommand
	}
	return options
}

func userDefaultsPath(home string) string {
	return filepath.Join(home, ".agents", ".env")
}

// DiscoverTechnologies finds selectable Markdown standards without a hard-coded list.
func DiscoverTechnologies(directory string) ([]Technology, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("discovering technology standards in %q: %w", directory, err)
	}
	var technologies []Technology
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("reading technology standard %q: %w", path, readErr)
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		technologies = append(technologies, Technology{Name: name, Title: firstHeading(contents, name), Path: path})
	}
	sort.Slice(technologies, func(i, j int) bool { return technologies[i].Name < technologies[j].Name })
	return technologies, nil
}

func firstHeading(contents []byte, fallback string) string {
	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return fallback
}

func resolveConfig(options Options, layers ...map[string]string) resolvedConfig {
	values := map[string]string{}
	for _, layer := range layers {
		for key, value := range layer {
			values[key] = value
		}
	}
	config := resolvedConfig{
		Harness: values[keyAgentHarness], SpecProvider: values[keySpecProvider], SpecModel: values[keySpecModel],
		BuildProvider: values[keyBuildProvider], BuildModel: values[keyBuildModel],
		AuditProvider: values[keyAuditProvider], AuditModel: values[keyAuditModel], InfraOwner: values[keyInfraOwner], InfraContract: values[keyInfraContract],
		Technologies: splitList(values[keyTechnologies]), SDLCRevision: options.SDLCRevision,
	}
	if value, ok := parseBool(values[keyInfraEnabled]); ok {
		config.InfraEnabled = &value
	}
	if options.Harness != "" {
		config.Harness = options.Harness
	}
	if options.SpecProvider != "" {
		config.SpecProvider = options.SpecProvider
	}
	if options.SpecModel != "" {
		config.SpecModel = options.SpecModel
	}
	if options.BuildProvider != "" {
		config.BuildProvider = options.BuildProvider
	}
	if options.BuildModel != "" {
		config.BuildModel = options.BuildModel
	}
	if options.AuditProvider != "" {
		config.AuditProvider = options.AuditProvider
	}
	if options.AuditModel != "" {
		config.AuditModel = options.AuditModel
	}
	if len(options.Technologies) != 0 {
		config.Technologies = options.Technologies
	}
	if options.InfraEnabled != nil {
		config.InfraEnabled = options.InfraEnabled
	}
	if options.InfraOwner != "" {
		config.InfraOwner = options.InfraOwner
	}
	if options.InfraContract != "" {
		config.InfraContract = options.InfraContract
	}
	return config
}

func projectSnapshot(config resolvedConfig) map[string]string {
	values := map[string]string{
		keyAgentHarness:  config.Harness,
		keySpecProvider:  config.SpecProvider,
		keySpecModel:     config.SpecModel,
		keyBuildProvider: config.BuildProvider,
		keyBuildModel:    config.BuildModel,
		keyAuditProvider: config.AuditProvider,
		keyAuditModel:    config.AuditModel,
		keyInfraOwner:    config.InfraOwner,
		keyInfraContract: config.InfraContract,
		keyTechnologies:  strings.Join(config.Technologies, ","),
	}
	if config.InfraEnabled != nil {
		values[keyInfraEnabled] = strconv.FormatBool(*config.InfraEnabled)
	}
	return values
}

func normalizeLegacyDelivery(values map[string]string) bool {
	changed := false
	if values[keySpecProvider] == "" && values[legacyDeliveryProvider] != "" {
		values[keySpecProvider] = values[legacyDeliveryProvider]
		changed = true
	}
	if values[keySpecModel] == "" && values[legacyDeliveryModel] != "" {
		values[keySpecModel] = values[legacyDeliveryModel]
		changed = true
	}
	delete(values, legacyDeliveryProvider)
	delete(values, legacyDeliveryModel)
	return changed
}

func splitList(value string) []string {
	var values []string
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, strings.ToUpper(trimmed))
		}
	}
	return values
}

func parseBool(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "1":
		return true, true
	case "false", "no", "0":
		return false, true
	default:
		return false, false
	}
}

func selectTechnologies(available []Technology, requested []string) ([]Technology, error) {
	byName := map[string]Technology{}
	for _, technology := range available {
		byName[strings.ToUpper(technology.Name)] = technology
	}
	seen := map[string]bool{}
	var selected []Technology
	for _, name := range requested {
		key := strings.ToUpper(strings.TrimSpace(name))
		technology, ok := byName[key]
		if !ok {
			return nil, fmt.Errorf("unknown technology standard %q", name)
		}
		if !seen[key] {
			selected = append(selected, technology)
			seen[key] = true
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Name < selected[j].Name })
	return selected, nil
}

func renderConstitution(sdlcRoot string, technologies []Technology, config resolvedConfig) []byte {
	var output strings.Builder
	output.WriteString("# [PROJECT_NAME] Constitution\n\n")
	output.WriteString("<!-- SDLC-GENERATED-BASELINE: do not remove or weaken generated clauses. -->\n\n")
	output.WriteString("## Engineering Standards\n\n")
	output.WriteString("This project MUST comply with the following canonical standards. The standards are referenced, not copied; load only those relevant to the current operation.\n\n")
	for _, item := range universalStandards {
		fmt.Fprintf(&output, "- **%s:** `%s`.\n", item.Subject, path.Join(sdlcRoot, item.Path))
	}
	for _, technology := range technologies {
		fmt.Fprintf(&output, "- **%s:** `%s`.\n", technology.Title, path.Join(sdlcRoot, "technologies", technology.Name+".md"))
	}
	output.WriteString("\nA deviation MUST name the standard, reason, risk, and approving authority. Silence is not a deviation.\n\n")
	if config.SDLCRevision == "" {
		output.WriteString("TODO(SDLC_REVISION): Pin the canonical SDLC release or Git revision before ratification.\n\n")
	} else {
		fmt.Fprintf(&output, "The adopted SDLC revision is `%s`.\n\n", config.SDLCRevision)
	}
	if config.InfraEnabled != nil && *config.InfraEnabled {
		output.WriteString("## External Infrastructure Ownership\n\n")
		fmt.Fprintf(&output, "%s owns deployment and runtime infrastructure. This project MUST remain deployable through and comply with `%s`. Application semantics remain project-owned; infrastructure mechanisms and obligations remain contract-owned. Conflicting historical deployment requirements MUST be classified against the current contract rather than silently retained.\n\n", config.InfraOwner, config.InfraContract)
	}
	output.WriteString("## Specification and Evidence\n\n")
	output.WriteString("No implementation may begin without a defined specification. Brownfield work MUST preserve lineage to previously implemented requirements and regression evidence. Project documentation and the active specification MUST be updated when delivered behaviour or ownership boundaries change.\n\n")
	output.WriteString("## Mandatory Independent Audits\n\n")
	output.WriteString("Each audit MUST run in a fresh agent context that did not author the artefact. It MUST emit the exact structured verdict required by its skill. Any finding, missing verdict, malformed verdict, or change to the audited artefact invalidates PASS. On FAIL, the author remediates and a fresh independent audit runs. The next stage MUST NOT begin until the required audit records PASS.\n\n")
	if config.AuditProvider != "" {
		fmt.Fprintf(&output, "The configured audit runtime is provider `%s`, model `%s`. If runtime configuration is absent, use the audit skill's preferred provider and model when available.\n\n", config.AuditProvider, config.AuditModel)
	}
	output.WriteString("1. Specification and clarification require `audit-spec` PASS before planning.\n")
	output.WriteString("2. Plan and design require `audit-design` PASS before test design and tasks.\n")
	output.WriteString("3. Test design and traceability require `audit-tests` PASS before implementation.\n")
	output.WriteString("4. Implementation requires `audit-code` PASS before completion or convergence.\n\n")
	output.WriteString("Record each audit name, auditor provider and model, artefact revision, exact verdict, findings, and superseding rerun in the active feature's `audits.md`. `speckit-analyze` is a consistency check and does not replace an independent audit.\n\n")
	output.WriteString("## Project-Specific Principles\n\n")
	output.WriteString("[ADD DURABLE PROJECT-SPECIFIC PRINCIPLES DERIVED FROM VERIFIED PROJECT DOCUMENTATION.]\n\n")
	output.WriteString("## Project Ownership and Architecture Boundaries\n\n")
	output.WriteString("[ADD DURABLE APPLICATION, DATA, INTEGRATION, AND OPERATIONAL OWNERSHIP BOUNDARIES.]\n\n")
	output.WriteString("## Governance\n\n")
	output.WriteString("This constitution governs project specifications, plans, tasks, implementation, and review. Amendments MUST preserve the generated baseline, explain compatibility and migration effects, and update the version and dates below.\n\n")
	output.WriteString("**Version**: [CONSTITUTION_VERSION] | **Ratified**: [RATIFICATION_DATE] | **Last Revised**: [LAST_AMENDED_DATE]\n")
	return []byte(output.String())
}

func ensureSpecKit(projectRoot string, config *resolvedConfig, reader *bufio.Reader, options Options) error {
	if info, err := os.Stat(filepath.Join(projectRoot, ".specify")); err == nil && info.IsDir() {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking Spec Kit project: %w", err)
	}
	accepted, err := promptYesNo(reader, options.Output, "Spec Kit is not initialized. Run specify init now? [yes/no]: ")
	if err != nil {
		return err
	}
	if !accepted {
		return errors.New("Spec Kit initialization declined; no files were changed")
	}
	if config.Harness == "" {
		config.Harness, err = promptRequired(reader, options.Output, "Spec Kit integration (codex, claude, or hermes): ")
		if err != nil {
			return err
		}
	}
	if err := validateHarness(config.Harness); err != nil {
		return err
	}
	arguments := []string{"init", "--here", "--force", "--non-interactive", "--integration", config.Harness, "--script", "sh"}
	return options.RunCommand("specify", arguments, projectRoot, options.Input, options.Output, options.ErrorOutput)
}

func ensurePreset(projectRoot, sdlcRoot string, options Options) error {
	destination := filepath.Join(projectRoot, ".specify", "presets", "sdlc-standards")
	preset := filepath.Join(destination, "preset.yml")
	source := filepath.Join(sdlcRoot, "presets", "sdlc-standards")
	if info, err := os.Stat(preset); err == nil && info.Mode().IsRegular() {
		equal, compareErr := directoriesEqual(source, destination)
		if compareErr != nil {
			return compareErr
		}
		if equal {
			return nil
		}
		backup := fmt.Sprintf("%s.%d.bak", destination, time.Now().UnixNano())
		if err := copyDirectory(destination, backup); err != nil {
			return fmt.Errorf("backing up changed SDLC preset: %w", err)
		}
		fmt.Fprintf(options.Output, "Updating changed SDLC preset; previous copy backed up at %s\n", backup)
		if err := options.RunCommand("specify", []string{"preset", "remove", "sdlc-standards"}, projectRoot, options.Input, options.Output, options.ErrorOutput); err != nil {
			return err
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking SDLC preset: %w", err)
	}
	return options.RunCommand("specify", []string{"preset", "add", "--dev", source}, projectRoot, options.Input, options.Output, options.ErrorOutput)
}

func directoriesEqual(source, destination string) (bool, error) {
	sourceFiles, err := directoryContents(source)
	if err != nil {
		return false, err
	}
	destinationFiles, err := directoryContents(destination)
	if err != nil {
		return false, err
	}
	if len(sourceFiles) != len(destinationFiles) {
		return false, nil
	}
	for relative, sourceContent := range sourceFiles {
		destinationContent, ok := destinationFiles[relative]
		if !ok || !bytes.Equal(sourceContent, destinationContent) {
			return false, nil
		}
	}
	return true, nil
}

func directoryContents(root string) (map[string][]byte, error) {
	contents := map[string][]byte{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() && path != root && filepath.Base(path) == ".composed" {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("preset path %q is not a regular file", path)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		contents[relative], err = os.ReadFile(path)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("reading preset directory %q: %w", root, err)
	}
	return contents, nil
}

func copyDirectory(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("preset path %q is not a regular file", path)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeAtomic(target, contents, info.Mode().Perm())
	})
}

func promptTechnologies(reader *bufio.Reader, output io.Writer, technologies []Technology) ([]string, error) {
	fmt.Fprintln(output, "Available technology standards:")
	for _, technology := range technologies {
		fmt.Fprintf(output, "- %s: %s\n", technology.Name, technology.Title)
	}
	fmt.Fprint(output, "Applicable technologies (comma-separated names): ")
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("reading technologies: %w", err)
	}
	return splitList(line), nil
}

func promptRequired(reader *bufio.Reader, output io.Writer, prompt string) (string, error) {
	fmt.Fprint(output, prompt)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("reading response: %w", err)
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return "", errors.New("a response is required")
	}
	return value, nil
}

func promptYesNo(reader *bufio.Reader, output io.Writer, prompt string) (bool, error) {
	value, err := promptRequired(reader, output, prompt)
	if err != nil {
		return false, err
	}
	parsed, ok := parseBool(value)
	if !ok {
		return false, fmt.Errorf("expected yes or no, got %q", value)
	}
	return parsed, nil
}

func validateHarness(harness string) error {
	switch strings.ToLower(strings.TrimSpace(harness)) {
	case "codex", "claude", "hermes":
		return nil
	default:
		return fmt.Errorf("unsupported agent harness %q; use codex, claude, or hermes", harness)
	}
}

func validateProviderModelPairs(config resolvedConfig) error {
	for _, pair := range []struct{ label, provider, model string }{
		{label: "specification", provider: config.SpecProvider, model: config.SpecModel},
		{label: "build", provider: config.BuildProvider, model: config.BuildModel},
		{label: "audit", provider: config.AuditProvider, model: config.AuditModel},
	} {
		if (pair.provider == "") != (pair.model == "") {
			return fmt.Errorf("%s provider and model must be configured together", pair.label)
		}
	}
	return nil
}

func launchConstitution(config resolvedConfig, projectRoot, templatePath string, options Options) error {
	prompt := constitutionPrompt(templatePath)
	harness := strings.ToLower(config.Harness)
	var arguments []string
	switch harness {
	case "codex":
		if config.SpecModel != "" {
			arguments = append(arguments, "--model", config.SpecModel)
		}
		arguments = append(arguments, prompt)
	case "claude":
		if config.SpecModel != "" {
			arguments = append(arguments, "--model", config.SpecModel)
		}
		arguments = append(arguments, prompt)
	case "hermes":
		arguments = []string{"chat", "--query", prompt, "--in", projectRoot}
		if config.SpecProvider != "" {
			arguments = append(arguments, "--provider", config.SpecProvider)
		}
		if config.SpecModel != "" {
			arguments = append(arguments, "--model", config.SpecModel)
		}
	}
	return options.RunCommand(harness, arguments, projectRoot, options.Input, options.Output, options.ErrorOutput)
}

func constitutionPrompt(templatePath string) string {
	return strings.Join([]string{
		"$speckit-constitution",
		"",
		"Create an unratified project constitution using the generated template at `" + templatePath + "` as an immutable baseline.",
		"",
		"This is a filtering exercise, not a summary of the project documentation. Read the project documentation as evidence, but do not restate its important requirements.",
		"",
		"Include a project-specific clause only when all of these are true:",
		"",
		"1. It applies across unrelated future features.",
		"2. It is expected to remain stable for years.",
		"3. Changing it would require a constitutional decision, not merely an approved feature specification or technical design.",
		"4. It defines project-wide authority, ownership, policy, or an invariant.",
		"5. It expresses a durable project-wide invariant supported by authoritative project documentation without reproducing that documentation's detailed requirements.",
		"",
		"Exclude:",
		"",
		"- feature requirements and acceptance criteria;",
		"- user journeys, actor permissions, approval sequences, and state transitions;",
		"- migration algorithms, schema procedures, commands, test gates, and validation procedures;",
		"- detailed architecture, component responsibilities, interfaces, and operational mechanisms; and",
		"- historical examples and unresolved feature or design questions.",
		"",
		"Project documentation remains authoritative for detailed requirements and design. The constitution may elevate a concise project-wide invariant supported by those documents. Do not copy the supporting feature behaviour, mechanisms, examples, or procedures. Importance alone does not make something constitutional.",
		"",
		"Add up to four concise project-specific principles. Zero is valid only when no evidenced project-wide invariant would require a constitutional amendment to change. The existence of an invariant in another authoritative document is not, by itself, a reason to omit it. Use one concise authority and ownership hierarchy. Record only explicit standards deviations and genuine ratification blockers. Do not repeat or expand the generated baseline.",
		"",
		"Make the authority hierarchy concern-specific. The human project owner or explicitly named human governance authority controls ratification, amendments, and deviations. The constitution and selected standards govern engineering and governance. Durable product vision and policy govern project purpose. Approved feature specifications govern observable behaviour. Approved architecture and design govern technical choices within those requirements. Operational, testing, migration, and user documentation govern their respective procedures and evidence. Code and tests record implemented state and evidence; they do not approve requirements. An external integration contract is authoritative only within its named ownership boundary. Do not place undifferentiated project documentation above approved specifications.",
		"",
		"The Governance section must name the human ratification and amendment authority, state whether any standards deviation is approved, and require compliance review to report applicable principles, deviations, and unresolved constitutional conflicts. Use a pre-1.0 version for an unratified draft and 1.0.0 for first ratification. After ratification, MAJOR removes or incompatibly redefines governance, MINOR adds or materially expands it, and PATCH clarifies it without changing meaning. List every unresolved ratification blocker; do not call one the sole blocker while another TODO remains.",
		"",
		"Before writing, test every proposed clause against this question:",
		"",
		"Would changing this require amending the constitution?",
		"",
		"If not, omit it.",
		"",
		"Produce only `.specify/memory/constitution.md`.",
	}, "\n")
}

func runCommand(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
	command := exec.Command(name, arguments...)
	command.Dir = directory
	command.Stdin, command.Stdout, command.Stderr = input, output, errorOutput
	if err := command.Run(); err != nil {
		return fmt.Errorf("running %s: %w", name, err)
	}
	return nil
}

func readManagedEnv(path string) (map[string]string, error) {
	values := map[string]string{}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) || path == "" {
		return values, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading SDLC configuration %q: %w", path, err)
	}
	defer file.Close()
	allowed := map[string]bool{
		legacyDeliveryProvider: true,
		legacyDeliveryModel:    true,
	}
	for _, key := range managedKeys {
		allowed[key] = true
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if found && allowed[key] {
			value = strings.TrimSpace(value)
			if strings.HasPrefix(value, "\"") {
				unquoted, unquoteErr := strconv.Unquote(value)
				if unquoteErr != nil {
					return nil, fmt.Errorf("reading SDLC configuration %q: invalid quoted value for %s", path, key)
				}
				value = unquoted
			} else {
				value = strings.Trim(value, "'")
			}
			values[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading SDLC configuration %q: %w", path, err)
	}
	return values, nil
}

func writeManagedEnv(path string, values map[string]string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading project environment %q: %w", path, err)
	}
	managed := map[string]bool{}
	for _, key := range managedKeys {
		managed[key] = true
	}
	managed[legacyDeliveryProvider] = true
	managed[legacyDeliveryModel] = true
	var kept []string
	for _, line := range strings.Split(strings.TrimSuffix(string(existing), "\n"), "\n") {
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "export "))
		key, _, found := strings.Cut(trimmed, "=")
		if found && managed[strings.TrimSpace(key)] {
			continue
		}
		if line != "" || len(kept) != 0 {
			kept = append(kept, line)
		}
	}
	if len(kept) != 0 && kept[len(kept)-1] != "" {
		kept = append(kept, "")
	}
	kept = append(kept, "# SDLC project configuration")
	for _, key := range managedKeys {
		if value, ok := values[key]; ok {
			kept = append(kept, key+"="+strconv.Quote(value))
		}
	}
	return writeAtomic(path, []byte(strings.Join(kept, "\n")+"\n"), 0o600)
}

func ensureEnvIgnored(path string) (bool, error) {
	contents, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("reading %q: %w", path, err)
	}
	for _, line := range strings.Split(string(contents), "\n") {
		if strings.TrimSpace(line) == ".env" {
			return false, nil
		}
	}
	updated := strings.TrimSuffix(string(contents), "\n")
	if updated != "" {
		updated += "\n"
	}
	updated += ".env\n"
	if err := writeAtomic(path, []byte(updated), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func writeAtomic(path string, contents []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %q: %w", filepath.Dir(path), err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".sdlc-*")
	if err != nil {
		return fmt.Errorf("creating temporary file for %q: %w", path, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err = temporary.Write(contents); err == nil {
		err = temporary.Chmod(mode)
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("writing %q: %w", path, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("installing %q: %w", path, err)
	}
	return nil
}

func printVariance(output io.Writer, before, after []byte) {
	fmt.Fprintln(output, "--- previous")
	fmt.Fprintln(output, "+++ rendered")
	for _, line := range strings.Split(strings.TrimSuffix(string(before), "\n"), "\n") {
		if len(before) != 0 {
			fmt.Fprintf(output, "- %s\n", line)
		}
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(after), "\n"), "\n") {
		if len(after) != 0 {
			fmt.Fprintf(output, "+ %s\n", line)
		}
	}
}
