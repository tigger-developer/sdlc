package projectinit

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
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
	"text/template"
	"time"
)

const (
	keyAgentHarness        = "SDLC_AGENT_HARNESS"
	keySpecProvider        = "SDLC_SPEC_PROVIDER"
	keySpecModel           = "SDLC_SPEC_MODEL"
	keyBuildProvider       = "SDLC_BUILD_PROVIDER"
	keyBuildModel          = "SDLC_BUILD_MODEL"
	keyAuditHarness        = "SDLC_AUDIT_HARNESS"
	keyAuditProvider       = "SDLC_AUDIT_PROVIDER"
	keyAuditModel          = "SDLC_AUDIT_MODEL"
	keyProjectType         = "SDLC_PROJECT_TYPE"
	keyInfraEnabled        = "SDLC_INFRA_ENABLED"
	keyInfraOwner          = "SDLC_INFRA_OWNER"
	keyInfraContract       = "SDLC_INFRA_CONTRACT"
	keyTechnologies        = "SDLC_TECHNOLOGIES"
	legacyDeliveryProvider = "SDLC_DELIVERY_PROVIDER"
	legacyDeliveryModel    = "SDLC_DELIVERY_MODEL"
	legacyACBlockPath      = "templates/project-init/legacy-acs-header.md"
	legacyACContractPath   = "templates/project-init/legacy-acs-header.json"
	constitutionLayoutPath = "templates/project-init/constitution-scaffold.md.tmpl"
	constitutionPromptPath = "prompts/project-init/constitution.md.tmpl"
	legacyACDocumentPath   = "docs/ACs.md"
)

var managedKeys = []string{
	keyAgentHarness,
	keySpecProvider,
	keySpecModel,
	keyBuildProvider,
	keyBuildModel,
	keyAuditHarness,
	keyAuditProvider,
	keyAuditModel,
	keyProjectType,
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
	{Path: "SECURITY.md", Subject: "Application security and vulnerability checking"},
	{Path: "PAIRING.md", Subject: "Paired development"},
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
	AuditHarness   string
	AuditProvider  string
	AuditModel     string
	ProjectType    string
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
	AuditHarness  string
	AuditProvider string
	AuditModel    string
	ProjectType   string
	Technologies  []string
	InfraEnabled  *bool
	InfraOwner    string
	InfraContract string
	SDLCRevision  string
}

type managedBlockContract struct {
	Marker     string `json:"marker"`
	Delimiter  string `json:"delimiter"`
	MarkerLine int    `json:"marker_line"`
}

type managedBlock struct {
	content    []byte
	marker     []byte
	delimiter  []byte
	markerLine int
	hash       [sha256.Size]byte
}

type constitutionTemplateData struct {
	Standards              []standard
	SDLCRevision           string
	ExternalInfrastructure bool
	InfrastructureOwner    string
	InfrastructureContract string
	ProjectType            string
}

var brownfieldMigrationPaths = []string{legacyACDocumentPath}

var requiredConstitutionSections = []string{
	"## Engineering Standards",
	"## Specification and Evidence",
	"## Mandatory Independent Audits",
}

// Run renders the deterministic constitution scaffold and, when it changes,
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
	if config.ProjectType != "" {
		if err := validateProjectType(config.ProjectType); err != nil {
			return err
		}
	}

	if err := ensureSpecKit(projectRoot, &config, reader, options); err != nil {
		return err
	}
	if err := ensurePreset(projectRoot, sdlcRoot, options); err != nil {
		return err
	}
	if config.ProjectType == "" {
		config.ProjectType, err = promptRequired(reader, options.Output, "Project type (greenfield or brownfield): ")
		if err != nil {
			return err
		}
		config.ProjectType = strings.ToLower(config.ProjectType)
		projectValues[keyProjectType] = config.ProjectType
		configChanged = true
	}
	if err := validateProjectType(config.ProjectType); err != nil {
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
		if current, exists := projectValues[key]; !exists || current != value {
			projectValues[key] = value
			configChanged = true
		}
	}

	referenceRoot := sdlcRoot
	if useCanonicalReference {
		referenceRoot = "~/.agents/sdlc"
	}
	constitutionLayout, err := readProjectInitResource(sdlcRoot, constitutionLayoutPath)
	if err != nil {
		return err
	}
	rendered, err := renderConstitution(constitutionLayout, referenceRoot, selected, config)
	if err != nil {
		return err
	}
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
	if config.ProjectType == "brownfield" {
		legacyBlock, blockErr := loadManagedBlock(sdlcRoot)
		if blockErr != nil {
			return blockErr
		}
		proceed, migrationErr := migrateBrownfieldDocuments(projectRoot, legacyBlock, reader, options)
		if migrationErr != nil {
			return migrationErr
		}
		if !proceed {
			return nil
		}
	}
	if templateChanged {
		if err := writeAtomic(target, rendered, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(options.Output, "Updated constitution scaffold: %s\n", target)
		if os.Getenv("VERBOSE") == "1" {
			printVariance(options.Output, current, rendered)
		}
	}
	if err := commitConstitutionScaffold(projectRoot, target, options); err != nil {
		return err
	}
	if !templateChanged && !configChanged {
		return nil
	}
	if options.NoLaunch {
		return nil
	}
	return launchConstitution(config, projectRoot, sdlcRoot, target, options)
}

func migrateBrownfieldDocuments(projectRoot string, block managedBlock, reader *bufio.Reader, options Options) (bool, error) {
	active, current, err := inspectBrownfieldMigration(projectRoot, block)
	if err != nil {
		return false, err
	}
	if !active {
		return true, nil
	}
	status, err := gitStatusForPaths(projectRoot, brownfieldMigrationPaths, options)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(status) != "" && !current {
		return false, fmt.Errorf("brownfield documentation migration overlaps existing changes:\n%s", strings.TrimSpace(status))
	}
	if !current {
		if _, err := ensureManagedPrefix(projectRoot, legacyACDocumentPath, block); err != nil {
			return false, err
		}
	}
	return reviewBrownfieldMigration(projectRoot, reader, current, options)
}

func inspectBrownfieldMigration(projectRoot string, block managedBlock) (bool, bool, error) {
	requirementsPath := filepath.Join(projectRoot, filepath.FromSlash(legacyACDocumentPath))
	requirementsExist, err := regularFileExists(requirementsPath)
	if err != nil {
		return false, false, err
	}
	if !requirementsExist {
		return false, false, nil
	}
	current, err := brownfieldMigrationCurrent(projectRoot, block)
	return true, current, err
}

func reviewBrownfieldMigration(projectRoot string, reader *bufio.Reader, current bool, options Options) (bool, error) {
	if current {
		return true, nil
	}
	if err := options.RunCommand("git", append([]string{"add", "--"}, brownfieldMigrationPaths...), projectRoot, strings.NewReader(""), options.Output, options.ErrorOutput); err != nil {
		return false, fmt.Errorf("staging brownfield documentation migration: %w", err)
	}
	fmt.Fprintln(options.Output, "Brownfield documentation migration variances:")
	diffArguments := append([]string{"diff", "--cached", "--no-ext-diff", "--find-renames", "--"}, brownfieldMigrationPaths...)
	if err := options.RunCommand("git", diffArguments, projectRoot, strings.NewReader(""), options.Output, options.ErrorOutput); err != nil {
		return false, fmt.Errorf("showing brownfield documentation migration: %w", err)
	}
	accepted, err := promptYesNo(reader, options.Output, "Commit these brownfield documentation migration changes? [yes/no]: ")
	if err != nil {
		return false, err
	}
	if !accepted {
		fmt.Fprintln(options.Output, "Brownfield documentation migration remains staged and constitution generation has stopped.")
		return false, nil
	}
	commitArguments := append([]string{"commit", "--only", "--quiet", "--message", "docs: mark legacy requirements authority", "--"}, brownfieldMigrationPaths...)
	if err := options.RunCommand("git", commitArguments, projectRoot, strings.NewReader(""), options.Output, options.ErrorOutput); err != nil {
		return false, fmt.Errorf("committing brownfield documentation migration: %w", err)
	}
	fmt.Fprintln(options.Output, "Committed brownfield documentation migration.")
	return true, nil
}

func brownfieldMigrationCurrent(projectRoot string, block managedBlock) (bool, error) {
	requirements, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(legacyACDocumentPath)))
	if err != nil {
		return false, fmt.Errorf("reading legacy requirements document %q: %w", legacyACDocumentPath, err)
	}
	_, changed, err := updateManagedPrefix(requirements, block)
	if err != nil {
		return false, fmt.Errorf("checking managed prefix in %q: %w", legacyACDocumentPath, err)
	}
	return !changed, nil
}

func ensureManagedPrefix(projectRoot, documentPath string, block managedBlock) (bool, error) {
	target := filepath.Join(projectRoot, filepath.FromSlash(documentPath))
	contents, err := os.ReadFile(target)
	if err != nil {
		return false, fmt.Errorf("reading authority document %q: %w", target, err)
	}
	updated, changed, err := updateManagedPrefix(contents, block)
	if err != nil {
		return false, fmt.Errorf("updating managed prefix in %q: %w", target, err)
	}
	if !changed {
		return false, nil
	}
	if err := writeAtomic(target, updated, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func readProjectInitResource(sdlcRoot, relativePath string) ([]byte, error) {
	clean := filepath.Clean(filepath.FromSlash(relativePath))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid project initializer resource path %q", relativePath)
	}
	target := filepath.Join(sdlcRoot, clean)
	contents, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("reading project initializer resource %q: %w", target, err)
	}
	return contents, nil
}

func loadManagedBlock(sdlcRoot string) (managedBlock, error) {
	content, err := readProjectInitResource(sdlcRoot, legacyACBlockPath)
	if err != nil {
		return managedBlock{}, err
	}
	contractContent, err := readProjectInitResource(sdlcRoot, legacyACContractPath)
	if err != nil {
		return managedBlock{}, err
	}
	var contract managedBlockContract
	if err := json.Unmarshal(contractContent, &contract); err != nil {
		return managedBlock{}, fmt.Errorf("parsing project initializer resource %q: %w", filepath.Join(sdlcRoot, filepath.FromSlash(legacyACContractPath)), err)
	}
	if contract.Marker == "" || contract.Delimiter == "" || contract.MarkerLine < 1 {
		return managedBlock{}, fmt.Errorf("invalid managed-block contract in %q", filepath.Join(sdlcRoot, filepath.FromSlash(legacyACContractPath)))
	}
	block := managedBlock{
		content:    content,
		marker:     []byte(contract.Marker),
		delimiter:  []byte(contract.Delimiter),
		markerLine: contract.MarkerLine,
		hash:       sha256.Sum256(content),
	}
	if !bytes.HasSuffix(content, []byte{'\n'}) {
		return managedBlock{}, fmt.Errorf("managed block %q must end with a newline", filepath.Join(sdlcRoot, filepath.FromSlash(legacyACBlockPath)))
	}
	if markerCount := bytes.Count(content, block.marker); markerCount != 1 {
		return managedBlock{}, fmt.Errorf("managed block %q contains %d markers; expected one", filepath.Join(sdlcRoot, filepath.FromSlash(legacyACBlockPath)), markerCount)
	}
	prefixEnd, err := managedPrefixEnd(content, block)
	if err != nil {
		return managedBlock{}, fmt.Errorf("validating managed block %q: %w", filepath.Join(sdlcRoot, filepath.FromSlash(legacyACBlockPath)), err)
	}
	if prefixEnd != len(content) {
		return managedBlock{}, fmt.Errorf("managed block %q contains content after its delimiter", filepath.Join(sdlcRoot, filepath.FromSlash(legacyACBlockPath)))
	}
	return block, nil
}

func updateManagedPrefix(contents []byte, block managedBlock) ([]byte, bool, error) {
	markerCount := bytes.Count(contents, block.marker)
	if markerCount == 0 {
		updated := make([]byte, 0, len(block.content)+1+len(contents))
		updated = append(updated, block.content...)
		updated = append(updated, '\n')
		updated = append(updated, contents...)
		return updated, true, nil
	}
	if markerCount != 1 {
		return nil, false, fmt.Errorf("managed marker occurs %d times", markerCount)
	}
	prefixEnd, err := managedPrefixEnd(contents, block)
	if err != nil {
		return nil, false, err
	}
	prefix := contents[:prefixEnd]
	if sha256.Sum256(prefix) == block.hash {
		return contents, false, nil
	}
	updated := make([]byte, 0, len(block.content)+len(contents)-prefixEnd)
	updated = append(updated, block.content...)
	updated = append(updated, contents[prefixEnd:]...)
	return updated, true, nil
}

func managedPrefixEnd(contents []byte, block managedBlock) (int, error) {
	markerIndex := bytes.Index(contents, block.marker)
	if markerIndex < 0 {
		return 0, errors.New("managed marker is absent")
	}
	markerLineStart := bytes.LastIndex(contents[:markerIndex], []byte{'\n'}) + 1
	markerLineEnd := bytes.IndexByte(contents[markerIndex:], '\n')
	if markerLineEnd < 0 {
		markerLineEnd = len(contents)
	} else {
		markerLineEnd += markerIndex
	}
	if !bytes.Equal(contents[markerLineStart:markerLineEnd], block.marker) {
		return 0, errors.New("managed marker must occupy its own line")
	}
	lineNumber := bytes.Count(contents[:markerLineStart], []byte{'\n'}) + 1
	if lineNumber != block.markerLine {
		return 0, fmt.Errorf("managed marker is on line %d, expected line %d", lineNumber, block.markerLine)
	}

	offset := 0
	for offset < len(contents) {
		lineEnd := bytes.IndexByte(contents[offset:], '\n')
		if lineEnd < 0 {
			lineEnd = len(contents)
			if bytes.Equal(contents[offset:lineEnd], block.delimiter) {
				if markerIndex >= lineEnd {
					return 0, errors.New("managed delimiter appears before the marker")
				}
				return lineEnd, nil
			}
			break
		}
		lineEnd += offset
		if bytes.Equal(contents[offset:lineEnd], block.delimiter) {
			if markerIndex >= lineEnd {
				return 0, errors.New("managed delimiter appears before the marker")
			}
			return lineEnd + 1, nil
		}
		offset = lineEnd + 1
	}
	return 0, errors.New("managed delimiter is absent")
}

func regularFileExists(target string) (bool, error) {
	info, err := os.Stat(target)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking %q: %w", target, err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("expected regular file at %q", target)
	}
	return true, nil
}

func gitStatusForPaths(projectRoot string, paths []string, options Options) (string, error) {
	var status bytes.Buffer
	arguments := append([]string{"status", "--porcelain=v1", "--untracked-files=all", "--"}, paths...)
	if err := options.RunCommand("git", arguments, projectRoot, strings.NewReader(""), &status, options.ErrorOutput); err != nil {
		return "", fmt.Errorf("checking brownfield documentation Git status: %w", err)
	}
	return status.String(), nil
}

func commitConstitutionScaffold(projectRoot, target string, options Options) error {
	relative, err := filepath.Rel(projectRoot, target)
	if err != nil {
		return fmt.Errorf("resolving constitution scaffold path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("constitution scaffold %q is outside project root %q", target, projectRoot)
	}
	relative = filepath.ToSlash(relative)

	var status bytes.Buffer
	statusArguments := []string{"status", "--porcelain=v1", "--untracked-files=all", "--", relative}
	if err := options.RunCommand("git", statusArguments, projectRoot, strings.NewReader(""), &status, options.ErrorOutput); err != nil {
		return fmt.Errorf("checking constitution scaffold Git status: %w", err)
	}
	if strings.TrimSpace(status.String()) == "" {
		return nil
	}
	if err := options.RunCommand("git", []string{"add", "--", relative}, projectRoot, strings.NewReader(""), options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("staging constitution scaffold: %w", err)
	}
	commitArguments := []string{"commit", "--only", "--quiet", "--message", "docs: update SDLC constitution scaffold", "--", relative}
	if err := options.RunCommand("git", commitArguments, projectRoot, strings.NewReader(""), options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("committing constitution scaffold: %w", err)
	}
	fmt.Fprintf(options.Output, "Committed constitution scaffold: %s\n", relative)
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

func resolveConfig(options Options, userValues, projectValues map[string]string) resolvedConfig {
	values := map[string]string{}
	for key, value := range userValues {
		if key != keyProjectType {
			values[key] = value
		}
	}
	for key, value := range projectValues {
		values[key] = value
	}
	config := resolvedConfig{
		Harness: values[keyAgentHarness], SpecProvider: values[keySpecProvider], SpecModel: values[keySpecModel],
		BuildProvider: values[keyBuildProvider], BuildModel: values[keyBuildModel],
		AuditHarness: values[keyAuditHarness], AuditProvider: values[keyAuditProvider], AuditModel: values[keyAuditModel], InfraOwner: values[keyInfraOwner], InfraContract: values[keyInfraContract],
		ProjectType: strings.ToLower(values[keyProjectType]), Technologies: splitList(values[keyTechnologies]), SDLCRevision: options.SDLCRevision,
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
	if options.AuditHarness != "" {
		config.AuditHarness = options.AuditHarness
	}
	if options.AuditProvider != "" {
		config.AuditProvider = options.AuditProvider
	}
	if options.AuditModel != "" {
		config.AuditModel = options.AuditModel
	}
	if options.ProjectType != "" {
		config.ProjectType = strings.ToLower(strings.TrimSpace(options.ProjectType))
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
		keyAuditHarness:  config.AuditHarness,
		keyAuditProvider: config.AuditProvider,
		keyAuditModel:    config.AuditModel,
		keyProjectType:   config.ProjectType,
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

func renderConstitution(layout []byte, sdlcRoot string, technologies []Technology, config resolvedConfig) ([]byte, error) {
	standards := make([]standard, 0, len(universalStandards)+len(technologies))
	for _, item := range universalStandards {
		standards = append(standards, standard{
			Path:    path.Join(sdlcRoot, item.Path),
			Subject: item.Subject,
		})
	}
	for _, technology := range technologies {
		standards = append(standards, standard{
			Path:    path.Join(sdlcRoot, "technologies", technology.Name+".md"),
			Subject: technology.Title,
		})
	}
	parsed, err := template.New("constitution-scaffold").Option("missingkey=error").Parse(string(layout))
	if err != nil {
		return nil, fmt.Errorf("parsing constitution scaffold template: %w", err)
	}
	data := constitutionTemplateData{
		Standards:              standards,
		SDLCRevision:           config.SDLCRevision,
		ExternalInfrastructure: config.InfraEnabled != nil && *config.InfraEnabled,
		InfrastructureOwner:    config.InfraOwner,
		InfrastructureContract: config.InfraContract,
		ProjectType:            config.ProjectType,
	}
	var output bytes.Buffer
	if err := parsed.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("rendering constitution scaffold template: %w", err)
	}
	return output.Bytes(), nil
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

func validateProjectType(projectType string) error {
	switch strings.ToLower(strings.TrimSpace(projectType)) {
	case "greenfield", "brownfield":
		return nil
	default:
		return fmt.Errorf("unsupported project type %q; use greenfield or brownfield", projectType)
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

func launchConstitution(config resolvedConfig, projectRoot, sdlcRoot, templatePath string, options Options) error {
	promptTemplate, err := readProjectInitResource(sdlcRoot, constitutionPromptPath)
	if err != nil {
		return err
	}
	prompt, err := renderConstitutionPrompt(promptTemplate, templatePath)
	if err != nil {
		return err
	}
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
	if err := options.RunCommand(harness, arguments, projectRoot, options.Input, options.Output, options.ErrorOutput); err != nil {
		return err
	}
	scaffold, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("reading rendered constitution scaffold %q: %w", templatePath, err)
	}
	return validateConstitutionCandidate(projectRoot, scaffold)
}

func validateConstitutionCandidate(projectRoot string, scaffold []byte) error {
	candidatePath := filepath.Join(projectRoot, ".specify", "memory", "constitution.md")
	candidate, err := os.ReadFile(candidatePath)
	if err != nil {
		return fmt.Errorf("reading generated constitution %q: %w", candidatePath, err)
	}
	for _, heading := range requiredConstitutionSections {
		required, err := markdownSection(scaffold, heading)
		if err != nil {
			return fmt.Errorf("reading required constitution governance: %w", err)
		}
		actual, err := markdownSection(candidate, heading)
		if err != nil {
			return fmt.Errorf("generated constitution failed shared-governance validation: %w", err)
		}
		if !strings.Contains(normalizeMarkdown(actual), normalizeMarkdown(required)) {
			return fmt.Errorf("generated constitution changed or omitted required shared-governance section %q", heading)
		}
	}
	return nil
}

func markdownSection(document []byte, heading string) (string, error) {
	lines := strings.Split(strings.ReplaceAll(string(document), "\r\n", "\n"), "\n")
	start := -1
	for index, line := range lines {
		if strings.TrimSpace(line) == heading {
			start = index
			break
		}
	}
	if start < 0 {
		return "", fmt.Errorf("missing section %q", heading)
	}
	end := len(lines)
	for index := start + 1; index < len(lines); index++ {
		if strings.HasPrefix(strings.TrimSpace(lines[index]), "## ") {
			end = index
			break
		}
	}
	return strings.Join(lines[start:end], "\n"), nil
}

func normalizeMarkdown(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func renderConstitutionPrompt(layout []byte, templatePath string) (string, error) {
	parsed, err := template.New("constitution-prompt").Option("missingkey=error").Parse(string(layout))
	if err != nil {
		return "", fmt.Errorf("parsing constitution prompt template: %w", err)
	}
	var output bytes.Buffer
	if err := parsed.Execute(&output, struct{ TemplatePath string }{TemplatePath: templatePath}); err != nil {
		return "", fmt.Errorf("rendering constitution prompt template: %w", err)
	}
	return output.String(), nil
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
