package projectinit

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/tigger-developer/sdlc/internal/configenv"
)

const (
	keyAgentHarness        = "SDLC_AGENT_HARNESS"
	keySpecHarness         = "SDLC_SPEC_HARNESS"
	keySpecProvider        = "SDLC_SPEC_PROVIDER"
	keySpecModel           = "SDLC_SPEC_MODEL"
	keyBuildHarness        = "SDLC_BUILD_HARNESS"
	keyBuildProvider       = "SDLC_BUILD_PROVIDER"
	keyBuildModel          = "SDLC_BUILD_MODEL"
	keyAuditHarness        = "SDLC_AUDIT_HARNESS"
	keyAuditProvider       = "SDLC_AUDIT_PROVIDER"
	keyAuditModel          = "SDLC_AUDIT_MODEL"
	keyBranchStrategy      = "SDLC_BRANCH_STRATEGY"
	keyProjectType         = "SDLC_PROJECT_TYPE"
	keyInfraRole           = "SDLC_INFRA_ROLE"
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
	migratedACDocumentPath = "docs/ACs.org"
	ticketMigrationPath    = "docs/ticket-migration.org"
	ticketManifestPath     = "docs/archive/migrated-tickets/manifest.json"
	environmentLoaderPath  = "libexec/load-sdlc-env.sh"
)

const legacyIssueProbeTimeout = 30 * time.Second

var universalStandards = []standard{
	{Path: "MAIN.md", Subject: "Universal engineering behaviour"},
	{Path: "ISSUES.md", Subject: "Specification and requirement quality"},
	{Path: "CODING.md", Subject: "Implementation and design"},
	{Path: "TESTING.md", Subject: "Testing and evidence"},
	{Path: "SECURITY.md", Subject: "Security and vulnerability checking"},
	{Path: "AUDITS.md", Subject: "Independent audits"},
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
	ProjectRoot                string
	SDLCRoot                   string
	UserConfigPath             string
	SDLCRevision               string
	NoLaunch                   bool
	UpdateConstitutionRevision bool
	OfferLegacyMigration       bool
	Input                      io.Reader
	Output                     io.Writer
	ErrorOutput                io.Writer
	LookupEnv                  func(string) (string, bool)
	Overrides                  map[string]string
	LoadEnvironment            func(string, string, []string) (map[string]string, error)
	RunCommand                 func(string, []string, string, io.Reader, io.Writer, io.Writer) error
	RunBoundedCommand          func(string, []string, string, io.Reader, io.Writer, io.Writer, time.Duration) error
}

type resolvedConfig struct {
	Harness        string
	SpecHarness    string
	SpecProvider   string
	SpecModel      string
	BuildHarness   string
	BuildProvider  string
	BuildModel     string
	AuditHarness   string
	AuditProvider  string
	AuditModel     string
	BranchStrategy string
	ProjectType    string
	Technologies   []string
	InfraRole      string
	InfraOwner     string
	InfraContract  string
	SDLCRevision   string
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
	InfrastructureRole     string
	InfrastructureOwner    string
	InfrastructureContract string
	BranchStrategy         string
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
	options.UserConfigPath, err = filepath.Abs(options.UserConfigPath)
	if err != nil {
		return fmt.Errorf("resolving user configuration path: %w", err)
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
	var constitutionState constitutionRevisionState
	if options.UpdateConstitutionRevision {
		if strings.TrimSpace(options.SDLCRevision) == "" {
			return errors.New("updating the constitution requires a versioned SDLC build")
		}
		if strings.ContainsAny(options.SDLCRevision, "`\r\n") {
			return fmt.Errorf("invalid SDLC revision %q", options.SDLCRevision)
		}
		constitutionState, err = readConstitutionRevision(projectRoot)
		if err != nil {
			return err
		}
	}

	schema, err := LoadConfigSchema(sdlcRoot)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(options.Input)
	projectConfigPath := filepath.Join(projectRoot, ".env")
	loaderPath := filepath.Join(sdlcRoot, filepath.FromSlash(environmentLoaderPath))
	environmentKeys := append(schema.ManagedKeys(), legacyDeliveryProvider, legacyDeliveryModel, keyInfraEnabled)
	userValues, err := options.LoadEnvironment(loaderPath, options.UserConfigPath, environmentKeys)
	if err != nil {
		return err
	}
	projectValues, err := options.LoadEnvironment(loaderPath, projectConfigPath, environmentKeys)
	if err != nil {
		return err
	}
	environmentValues := readManagedProcessEnvironment(schema, options.LookupEnv)
	normalizeLegacyDelivery(userValues)
	normalizeLegacyInfrastructure(userValues)
	legacyProjectConfig := normalizeLegacyDelivery(projectValues)
	legacyProjectConfig = normalizeLegacyInfrastructure(projectValues) || legacyProjectConfig
	normalizeLegacyDelivery(environmentValues)
	normalizeLegacyInfrastructure(environmentValues)
	cliValues := optionsOverrides(options)
	values := resolveConfigValues(schema, cliValues, userValues, projectValues, environmentValues)
	config := configFromValues(values, options.SDLCRevision)
	configChanged := legacyProjectConfig
	technologies, err := DiscoverTechnologies(filepath.Join(sdlcRoot, "technologies"))
	if err != nil {
		return err
	}
	promptedFields, err := completeConfig(schema, values, reader, options.Output, technologies)
	if err != nil {
		return err
	}
	configChanged = configChanged || len(promptedFields) != 0
	config = configFromValues(values, options.SDLCRevision)
	if err := validateConfig(schema, values, technologies); err != nil {
		return err
	}
	config = configFromValues(values, options.SDLCRevision)
	if options.OfferLegacyMigration {
		proceed, err := offerLegacyMigration(config, projectRoot, reader, options)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}
	}
	if err := ensureSpecKit(projectRoot, &config, reader, options); err != nil {
		return err
	}
	if err := ensurePreset(projectRoot, sdlcRoot, options); err != nil {
		return err
	}
	selected, err := selectTechnologies(technologies, config.Technologies)
	if err != nil {
		return err
	}
	for _, field := range schema.Fields {
		_, cliSelection := cliValues[field.Key]
		_, environmentSelection := environmentValues[field.Key]
		promptSelection := promptedFields[field.Key]
		if !field.Persist || (!cliSelection && !environmentSelection && !promptSelection) {
			continue
		}
		value := values[field.Key]
		if current, exists := projectValues[field.Key]; !exists || current != value {
			projectValues[field.Key] = value
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
		if err := writeManagedEnv(projectConfigPath, projectValues, schema); err != nil {
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
		if err := validateBrownfieldLedger(projectRoot); err != nil {
			return err
		}
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
	constitutionChanged := false
	if options.UpdateConstitutionRevision {
		var previous string
		constitutionChanged, previous, err = updateConstitutionRevision(constitutionState, options.SDLCRevision)
		if err != nil {
			return err
		}
		if constitutionChanged {
			fmt.Fprintf(options.Output, "Updated adopted SDLC revision: %s -> %s\n", previous, options.SDLCRevision)
		}
	}
	if !templateChanged && !configChanged && !constitutionChanged {
		return nil
	}
	if options.NoLaunch {
		return nil
	}
	return launchConstitution(config, projectRoot, sdlcRoot, target, options)
}

type legacyIssueSummary struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

func offerLegacyMigration(config resolvedConfig, projectRoot string, reader *bufio.Reader, options Options) (bool, error) {
	if config.ProjectType != "brownfield" {
		return true, nil
	}
	legacyLedger, err := regularFileExists(filepath.Join(projectRoot, filepath.FromSlash(legacyACDocumentPath)))
	if err != nil {
		return false, err
	}
	migratedLedger, err := regularFileExists(filepath.Join(projectRoot, filepath.FromSlash(migratedACDocumentPath)))
	if err != nil {
		return false, err
	}
	if !legacyLedger || migratedLedger {
		return true, nil
	}
	if options.NoLaunch {
		fmt.Fprintln(options.Output, "SDLC v1 migration is required; --no-launch prevents invoking the migration skill.")
		return false, nil
	}

	issues, err := listOpenLegacyIssues(projectRoot, options)
	if err != nil {
		return false, fmt.Errorf("checking open GitHub issues before SDLC v1 migration: %w", err)
	}
	if len(issues) == 0 {
		fmt.Fprintln(options.Output, "Detected an SDLC v1 acceptance-criteria ledger; GitHub reports no open issues.")
	} else {
		fmt.Fprintf(options.Output, "Detected an SDLC v1 acceptance-criteria ledger and open GitHub issue #%d - %s.\n", issues[0].Number, issues[0].Title)
	}
	accepted, err := promptYesNo(reader, options.Output, "Run $migrate-legacy-acs-to-sdlc-v1 before project initialization? [yes/no]: ")
	if err != nil {
		return false, err
	}
	if !accepted {
		fmt.Fprintln(options.Output, "Legacy migration declined; project initialization stopped before changes.")
		return false, nil
	}
	prompt := "Invoke $migrate-legacy-acs-to-sdlc-v1 for this project. Complete the skill before returning."
	if err := runConfiguredHarness(config.AuditHarness, config.AuditProvider, config.AuditModel, prompt, projectRoot, options); err != nil {
		return false, fmt.Errorf("running legacy-ticket migration skill: %w", err)
	}
	remaining, err := listOpenLegacyIssues(projectRoot, options)
	if err != nil {
		return false, fmt.Errorf("verifying legacy issue closure: %w", err)
	}
	if len(remaining) != 0 {
		return false, fmt.Errorf("legacy migration left open GitHub issue #%d - %s", remaining[0].Number, remaining[0].Title)
	}
	legacyLedger, err = regularFileExists(filepath.Join(projectRoot, filepath.FromSlash(legacyACDocumentPath)))
	if err != nil {
		return false, err
	}
	migratedLedger, err = regularFileExists(filepath.Join(projectRoot, filepath.FromSlash(migratedACDocumentPath)))
	if err != nil {
		return false, err
	}
	if legacyLedger || !migratedLedger {
		return false, fmt.Errorf("legacy migration must replace %s with %s before initialization", legacyACDocumentPath, migratedACDocumentPath)
	}
	return true, nil
}

func listOpenLegacyIssues(projectRoot string, options Options) ([]legacyIssueSummary, error) {
	var output bytes.Buffer
	// This is an existence probe, not the migration inventory. The migration
	// skill owns complete issue and comment pagination; one returned issue proves
	// that open work remains, while an empty result proves there is none.
	arguments := []string{"issue", "list", "--state", "open", "--limit", "1", "--json", "number,title"}
	if err := options.RunBoundedCommand("gh", arguments, projectRoot, strings.NewReader(""), &output, options.ErrorOutput, legacyIssueProbeTimeout); err != nil {
		return nil, err
	}
	var issues []legacyIssueSummary
	if err := json.Unmarshal(output.Bytes(), &issues); err != nil {
		return nil, fmt.Errorf("parsing gh issue list output: %w", err)
	}
	return issues, nil
}

type constitutionRevisionState struct {
	path     string
	revision string
	content  []byte
}

func updateConstitutionRevision(state constitutionRevisionState, revision string) (bool, string, error) {
	if state.revision == revision {
		return false, state.revision, nil
	}
	current, err := os.ReadFile(state.path)
	if err != nil {
		return false, "", fmt.Errorf("rereading project constitution %q: %w", state.path, err)
	}
	if !bytes.Equal(current, state.content) {
		return false, "", fmt.Errorf("project constitution %q changed after validation; refusing to overwrite it", state.path)
	}
	lines := strings.Split(string(state.content), "\n")
	for index, line := range lines {
		if adoptedRevisionFromLine(line) == state.revision {
			lines[index] = "The adopted SDLC revision is `" + revision + "`."
			break
		}
	}
	info, err := os.Stat(state.path)
	if err != nil {
		return false, "", fmt.Errorf("inspecting project constitution %q: %w", state.path, err)
	}
	if err := writeAtomic(state.path, []byte(strings.Join(lines, "\n")), info.Mode().Perm()); err != nil {
		return false, "", err
	}
	return true, state.revision, nil
}

func readConstitutionRevision(projectRoot string) (constitutionRevisionState, error) {
	path := filepath.Join(projectRoot, ".specify", "memory", "constitution.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return constitutionRevisionState{}, fmt.Errorf("reading project constitution %q: %w", path, err)
	}
	matchCount := 0
	revision := ""
	for _, line := range strings.Split(string(content), "\n") {
		candidate := adoptedRevisionFromLine(line)
		if candidate == "" {
			continue
		}
		matchCount++
		revision = candidate
	}
	if matchCount != 1 {
		return constitutionRevisionState{}, fmt.Errorf("project constitution %q must contain exactly one adopted SDLC revision", path)
	}
	return constitutionRevisionState{path: path, revision: revision, content: content}, nil
}

func adoptedRevisionFromLine(line string) string {
	const prefix = "The adopted SDLC revision is `"
	const suffix = "`."
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix)
}

func validateBrownfieldLedger(projectRoot string) error {
	markdownExists, err := regularFileExists(filepath.Join(projectRoot, filepath.FromSlash(legacyACDocumentPath)))
	if err != nil {
		return err
	}
	orgExists, err := regularFileExists(filepath.Join(projectRoot, filepath.FromSlash(migratedACDocumentPath)))
	if err != nil {
		return err
	}
	if markdownExists && orgExists {
		return fmt.Errorf("brownfield project contains both %s and %s; retain only the canonical Org ledger", migratedACDocumentPath, legacyACDocumentPath)
	}

	migrationIndexExists, err := regularFileExists(filepath.Join(projectRoot, filepath.FromSlash(ticketMigrationPath)))
	if err != nil {
		return err
	}
	migrationManifestExists, err := regularFileExists(filepath.Join(projectRoot, filepath.FromSlash(ticketManifestPath)))
	if err != nil {
		return err
	}
	if (migrationIndexExists || migrationManifestExists) && !orgExists {
		return fmt.Errorf("legacy ticket-migration artefacts exist but %s is missing; invoke $convert-migrated-acs-to-org before project initialization", migratedACDocumentPath)
	}
	return nil
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
	if options.LookupEnv == nil {
		options.LookupEnv = os.LookupEnv
	}
	if options.LoadEnvironment == nil {
		options.LoadEnvironment = configenv.Load
	}
	if options.RunCommand == nil {
		options.RunCommand = runCommand
	}
	if options.RunBoundedCommand == nil {
		options.RunBoundedCommand = runCommandWithTimeout
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

func optionsOverrides(options Options) map[string]string {
	values := map[string]string{}
	for key, value := range options.Overrides {
		if strings.TrimSpace(value) != "" {
			values[key] = value
		}
	}
	return values
}

func resolveConfigValues(schema ConfigSchema, cli, user, project, environment map[string]string) map[string]string {
	sources := map[string]map[string]string{
		"cli": cli, "environment": environment, "project": project, "user": user,
	}
	values := map[string]string{}
	for _, field := range schema.Fields {
		for _, sourceName := range schema.Precedence {
			if sourceName == "fallback" || (sourceName == "user" && !field.AllowUserDefault) {
				continue
			}
			if value := strings.TrimSpace(sources[sourceName][field.Key]); value != "" {
				values[field.Key] = value
				break
			}
		}
	}
	applyFallbacks(schema, values)
	return values
}

func applyFallbacks(schema ConfigSchema, values map[string]string) {
	for range schema.Fields {
		for _, field := range schema.Fields {
			if values[field.Key] == "" && field.Fallback != "" && values[field.Fallback] != "" {
				values[field.Key] = values[field.Fallback]
			}
		}
	}
	for _, field := range schema.Fields {
		if values[field.Key] == "" && field.Default != "" {
			values[field.Key] = field.Default
		}
	}
}

func configFromValues(values map[string]string, revision string) resolvedConfig {
	return resolvedConfig{
		Harness: values[keyAgentHarness], SpecHarness: values[keySpecHarness], SpecProvider: values[keySpecProvider], SpecModel: values[keySpecModel],
		BuildHarness: values[keyBuildHarness], BuildProvider: values[keyBuildProvider], BuildModel: values[keyBuildModel],
		AuditHarness: values[keyAuditHarness], AuditProvider: values[keyAuditProvider], AuditModel: values[keyAuditModel],
		BranchStrategy: strings.ToLower(values[keyBranchStrategy]),
		ProjectType:    strings.ToLower(values[keyProjectType]), Technologies: splitList(values[keyTechnologies]),
		InfraRole: strings.ToLower(values[keyInfraRole]), InfraOwner: values[keyInfraOwner], InfraContract: values[keyInfraContract],
		SDLCRevision: revision,
	}
}

func readManagedProcessEnvironment(schema ConfigSchema, lookup func(string) (string, bool)) map[string]string {
	values := map[string]string{}
	if lookup == nil {
		return values
	}
	keys := append(schema.ManagedKeys(), legacyDeliveryProvider, legacyDeliveryModel, keyInfraEnabled)
	for _, key := range keys {
		if value, ok := lookup(key); ok {
			values[key] = value
		}
	}
	return values
}

func normalizeLegacyInfrastructure(values map[string]string) bool {
	legacy, exists := values[keyInfraEnabled]
	if !exists {
		return false
	}
	if values[keyInfraRole] == "" {
		if enabled, ok := parseBool(legacy); ok {
			if enabled {
				values[keyInfraRole] = "consumer"
			} else {
				values[keyInfraRole] = "none"
			}
		}
	}
	delete(values, keyInfraEnabled)
	return true
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
		InfrastructureRole:     config.InfraRole,
		InfrastructureOwner:    config.InfraOwner,
		InfrastructureContract: config.InfraContract,
		BranchStrategy:         config.BranchStrategy,
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
	if config.SpecHarness == "" {
		config.SpecHarness, err = promptRequired(reader, options.Output, "Spec Kit integration (codex, claude, or hermes): ")
		if err != nil {
			return err
		}
	}
	arguments := []string{"init", "--here", "--force", "--non-interactive", "--integration", config.SpecHarness, "--script", "sh"}
	return options.RunCommand("specify", arguments, projectRoot, options.Input, options.Output, options.ErrorOutput)
}

func ensurePreset(projectRoot, sdlcRoot string, options Options) error {
	destination := filepath.Join(projectRoot, ".specify", "presets", "sdlc-standards")
	preset := filepath.Join(destination, "preset.yml")
	source := filepath.Join(sdlcRoot, "presets", "sdlc-standards")
	backup := ""
	cleanupBackup := func() error { return nil }
	if info, err := os.Stat(preset); err == nil && info.Mode().IsRegular() {
		equal, compareErr := directoriesEqual(source, destination)
		if compareErr != nil {
			return compareErr
		}
		if equal && !options.UpdateConstitutionRevision {
			return nil
		}
		if equal {
			backupRoot, err := os.MkdirTemp("", "sdlc-preset-backup-")
			if err != nil {
				return fmt.Errorf("creating temporary SDLC preset backup: %w", err)
			}
			cleanupBackup = func() error { return os.RemoveAll(backupRoot) }
			backup = filepath.Join(backupRoot, "sdlc-standards")
		} else {
			backup = fmt.Sprintf("%s.%d.bak", destination, time.Now().UnixNano())
		}
		if err := copyDirectory(destination, backup); err != nil {
			return errors.Join(
				fmt.Errorf("backing up SDLC preset before recomposition: %w", err),
				cleanupBackup(),
			)
		}
		if !equal {
			fmt.Fprintf(options.Output, "Updating changed SDLC preset; previous copy backed up at %s\n", backup)
		}
		if err := options.RunCommand("specify", []string{"preset", "remove", "sdlc-standards"}, projectRoot, options.Input, options.Output, options.ErrorOutput); err != nil {
			return errors.Join(err, cleanupBackup())
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking SDLC preset: %w", err)
	}
	if err := options.RunCommand("specify", []string{"preset", "add", "--dev", source}, projectRoot, options.Input, options.Output, options.ErrorOutput); err != nil {
		if backup == "" {
			return err
		}
		removeErr := os.RemoveAll(destination)
		restoreErr := copyDirectory(backup, destination)
		cleanupErr := cleanupBackup()
		if removeErr != nil || restoreErr != nil || cleanupErr != nil {
			return errors.Join(err, removeErr, restoreErr, cleanupErr)
		}
		return err
	}
	if err := cleanupBackup(); err != nil {
		return fmt.Errorf("removing temporary SDLC preset backup: %w", err)
	}
	return nil
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

func completeConfig(schema ConfigSchema, values map[string]string, reader *bufio.Reader, output io.Writer, technologies []Technology) (map[string]bool, error) {
	prompted := map[string]bool{}
	for _, field := range schema.Fields {
		if values[field.Key] == "" && field.Fallback != "" {
			values[field.Key] = values[field.Fallback]
		}
		if values[field.Key] != "" || field.Prompt == "" || !conditionApplies(field.RequiredWhen, values) {
			continue
		}
		var value string
		var err error
		switch field.Type {
		case "choice":
			value, err = promptChoice(reader, output, field.Prompt, field.Choices)
		case "multi-choice":
			value, err = promptMultipleChoices(reader, output, field, technologies)
		case "string", "duration":
			value, err = promptRequired(reader, output, strings.TrimSpace(field.Prompt)+" ")
		}
		if err != nil {
			return nil, err
		}
		values[field.Key] = value
		prompted[field.Key] = true
	}
	applyFallbacks(schema, values)
	return prompted, nil
}

func conditionApplies(condition *RequiredWhen, values map[string]string) bool {
	if condition == nil {
		return true
	}
	for _, expected := range condition.Values {
		if strings.EqualFold(values[condition.Field], expected) {
			return true
		}
	}
	return false
}

func promptMultipleChoices(reader *bufio.Reader, output io.Writer, field ConfigField, technologies []Technology) (string, error) {
	choices := append([]string(nil), field.Choices...)
	titles := map[string]string{}
	if field.ChoicesFrom == "technologies" {
		for _, technology := range technologies {
			choices = append(choices, technology.Name)
			titles[technology.Name] = technology.Title
		}
	}
	fmt.Fprintln(output, field.Prompt)
	for index, choice := range choices {
		if title := titles[choice]; title != "" {
			fmt.Fprintf(output, "%d. %s: %s\n", index+1, choice, title)
		} else {
			fmt.Fprintf(output, "%d. %s\n", index+1, choice)
		}
	}
	fmt.Fprint(output, "Selections (comma-separated numbers; blank for none): ")
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("reading selections for %s: %w", field.Key, err)
	}
	if strings.TrimSpace(line) == "" {
		return "", nil
	}
	seen := map[int]bool{}
	selected := make([]string, 0)
	for _, part := range strings.Split(line, ",") {
		selection, parseErr := strconv.Atoi(strings.TrimSpace(part))
		if parseErr != nil || selection < 1 || selection > len(choices) {
			return "", fmt.Errorf("expected comma-separated numbers from 1 to %d, got %q", len(choices), strings.TrimSpace(line))
		}
		if !seen[selection] {
			selected = append(selected, choices[selection-1])
			seen[selection] = true
		}
	}
	return strings.Join(selected, ","), nil
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

func promptChoice(reader *bufio.Reader, output io.Writer, prompt string, choices []string) (string, error) {
	if len(choices) == 0 {
		return "", errors.New("at least one choice is required")
	}
	fmt.Fprintln(output, prompt)
	for index, choice := range choices {
		fmt.Fprintf(output, "%d. %s\n", index+1, choice)
	}
	value, err := promptRequired(reader, output, "Selection: ")
	if err != nil {
		return "", err
	}
	selection, err := strconv.Atoi(value)
	if err != nil || selection < 1 || selection > len(choices) {
		return "", fmt.Errorf("expected a number from 1 to %d, got %q", len(choices), value)
	}
	return choices[selection-1], nil
}

func validateConfig(schema ConfigSchema, values map[string]string, technologies []Technology) error {
	for _, field := range schema.Fields {
		value := strings.TrimSpace(values[field.Key])
		if value == "" {
			if (field.Type == "choice" && field.Prompt != "") || (field.RequiredWhen != nil && conditionApplies(field.RequiredWhen, values)) {
				return fmt.Errorf("%s is required", field.Key)
			}
			continue
		}
		switch field.Type {
		case "choice":
			canonical, ok := canonicalChoice(field, value)
			if !ok {
				return fmt.Errorf("%s must be one of %s, got %q", field.Key, strings.Join(field.Choices, ", "), value)
			}
			values[field.Key] = canonical
		case "multi-choice":
			if field.ChoicesFrom == "technologies" {
				if _, err := selectTechnologies(technologies, splitList(value)); err != nil {
					return err
				}
			}
		case "duration":
			if err := validateWholeSecondDuration(field.Key, value); err != nil {
				return err
			}
		}
	}
	for _, phase := range schema.Phases {
		if !slices.Contains(phase.ProviderHarnesses, values[phase.Harness]) {
			continue
		}
		provider := strings.TrimSpace(values[phase.Provider])
		model := strings.TrimSpace(values[phase.Model])
		if (provider == "") != (model == "") {
			return fmt.Errorf("%s fields %s and %s must be configured together for %s", phase.Name, phase.Provider, phase.Model, values[phase.Harness])
		}
	}
	return nil
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

func launchConstitution(config resolvedConfig, projectRoot, sdlcRoot, templatePath string, options Options) error {
	promptTemplate, err := readProjectInitResource(sdlcRoot, constitutionPromptPath)
	if err != nil {
		return err
	}
	prompt, err := renderConstitutionPrompt(promptTemplate, templatePath)
	if err != nil {
		return err
	}
	if err := runSpecHarness(config, prompt, projectRoot, options); err != nil {
		return err
	}
	scaffold, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("reading rendered constitution scaffold %q: %w", templatePath, err)
	}
	return validateConstitutionCandidate(projectRoot, scaffold)
}

func runSpecHarness(config resolvedConfig, prompt, projectRoot string, options Options) error {
	return runConfiguredHarness(config.SpecHarness, config.SpecProvider, config.SpecModel, prompt, projectRoot, options)
}

func runConfiguredHarness(harness, provider, model, prompt, projectRoot string, options Options) error {
	harness = strings.ToLower(harness)
	var arguments []string
	switch harness {
	case "codex":
		if model != "" {
			arguments = append(arguments, "--model", model)
		}
		arguments = append(arguments, prompt)
	case "claude":
		if model != "" {
			arguments = append(arguments, "--model", model)
		}
		arguments = append(arguments, prompt)
	case "hermes":
		arguments = []string{"chat", "--query", prompt, "--in", projectRoot}
		if provider != "" {
			arguments = append(arguments, "--provider", provider)
		}
		if model != "" {
			arguments = append(arguments, "--model", model)
		}
	default:
		return fmt.Errorf("unsupported agent harness %q", harness)
	}
	if err := options.RunCommand(harness, arguments, projectRoot, options.Input, options.Output, options.ErrorOutput); err != nil {
		return err
	}
	return nil
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

func runCommandWithTimeout(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = directory
	command.Stdin, command.Stdout, command.Stderr = input, output, errorOutput
	if err := command.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("running %s timed out after %s: %w", name, timeout, context.DeadlineExceeded)
		}
		return fmt.Errorf("running %s: %w", name, err)
	}
	return nil
}

func writeManagedEnv(path string, values map[string]string, schema ConfigSchema) error {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading project environment %q: %w", path, err)
	}
	managed := map[string]bool{}
	for _, key := range schema.ManagedKeys() {
		managed[key] = true
	}
	managed[legacyDeliveryProvider] = true
	managed[legacyDeliveryModel] = true
	managed[keyInfraEnabled] = true
	var kept []string
	for _, line := range strings.Split(strings.TrimSuffix(string(existing), "\n"), "\n") {
		if strings.TrimSpace(line) == "# SDLC project configuration" {
			continue
		}
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
	for _, field := range schema.Fields {
		if value, ok := values[field.Key]; ok && field.Persist {
			kept = append(kept, field.Key+"="+shellQuote(value))
		}
	}
	return writeAtomic(path, []byte(strings.Join(kept, "\n")+"\n"), 0o600)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
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
