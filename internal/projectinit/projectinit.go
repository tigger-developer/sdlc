package projectinit

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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tigger-developer/sdlc/internal/configenv"
	"gopkg.in/yaml.v3"
)

const projectProfilePath = ".sdlc/project.yaml"

var historicalIdentifierPattern = regexp.MustCompile(`(?m)(?:^|[^A-Za-z0-9])(?:W|AC|RT|UT|OT)0*([0-9]+)(?:[.]|\b)|#0*([0-9]+)\b`)

// Technology is one automatically discovered technology standard.
type Technology struct {
	Name string
	Path string
}

// Options controls one once-only project initialization.
type Options struct {
	ProjectRoot          string
	SDLCRoot             string
	GlobalConfigPath     string
	OverrideGlobalConfig bool
	SkipAgentScans       bool
	Overrides            map[string]string
	Input                io.Reader
	Output               io.Writer
	ErrorOutput          io.Writer
	Now                  func() time.Time
	RunCommand           func(string, []string, string, io.Reader, io.Writer, io.Writer) error
	inputReader          *bufio.Reader
}

type projectGeneration struct {
	values     map[string]string
	explicit   map[string]bool
	source     string
	base       string
	archive    string
	migration  string
	date       string
	legacyV2   []string
	migratedV2 []migratedWorkCandidate
}

type promptChoice struct {
	Label    string
	Selected bool
}

// Run initializes or migrates one project to SDLC v3.
func Run(options Options) error {
	options = defaultOptions(options)
	projectRoot, err := filepath.Abs(options.ProjectRoot)
	if err != nil {
		return fmt.Errorf("resolving project root: %w", err)
	}
	sdlcRoot, err := filepath.Abs(options.SDLCRoot)
	if err != nil {
		return fmt.Errorf("resolving SDLC root: %w", err)
	}
	if err := ensureGitRepository(options, projectRoot); err != nil {
		return err
	}
	workspace := initializationWorkspacePath(projectRoot)
	resume, err := interruptedMigration(options, projectRoot, workspace)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(projectRoot, projectProfilePath)); err == nil && resume == nil {
		return fmt.Errorf("%s already exists; sdlc-init runs exactly once", projectProfilePath)
	} else if !errors.Is(err, os.ErrNotExist) {
		if err != nil {
			return fmt.Errorf("checking project profile: %w", err)
		}
	}
	if resume == nil {
		if err := requireCleanWorktree(options, projectRoot); err != nil {
			return err
		}
		if err := ensureInitialCommit(options, projectRoot); err != nil {
			return err
		}
	}

	source := ""
	if resume != nil {
		source = resume.source
		fmt.Fprintf(options.Output, "Resuming interrupted SDLC v3 initialization on %s.\n", resume.migration)
	} else {
		source, err = DetectSource(projectRoot)
		if err != nil {
			return err
		}
	}
	runTicketMigration := false
	if source == "v1" && resume == nil {
		runTicketMigration, err = chooseTicketMigration(options, projectRoot)
		if err != nil {
			return err
		}
	}

	schema, err := LoadConfigSchema(sdlcRoot)
	if err != nil {
		return err
	}
	technologies, err := DiscoverTechnologies(filepath.Join(sdlcRoot, "technologies"))
	if err != nil {
		return err
	}
	if err := validateTechnologyDetectionStandards(schema.TechnologyDetection, technologies); err != nil {
		return fmt.Errorf("validating technology detection against installed standards: %w", err)
	}
	globalPath := options.GlobalConfigPath
	if globalPath == "" {
		userHome, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return fmt.Errorf("resolving user home: %w", homeErr)
		}
		globalPath = filepath.Join(userHome, ".agents", "sdlc.yaml")
	}
	global, err := readYAMLConfig(globalPath)
	if err != nil {
		return err
	}
	legacy, err := loadLegacyEnvironment(sdlcRoot, projectRoot, schema)
	if err != nil {
		return err
	}
	legacy = normalizeLegacyConfiguration(legacy, options.ErrorOutput)
	technologyAssessment := assessProjectTechnologies(options, schema, technologies, global, legacy, projectRoot)
	values, explicit, err := resolveConfiguration(options, schema, technologies, global, legacy, technologyAssessment)
	if err != nil {
		return err
	}

	var base, archive, migration, date string
	if resume != nil {
		base, archive, migration, date = resume.base, resume.archive, resume.migration, resume.date
	} else {
		base, err = commandOutput(options, projectRoot, "git", "branch", "--show-current")
		if err != nil || strings.TrimSpace(base) == "" {
			return errors.New("the project must be on a named Git branch before initialization")
		}
		base = strings.TrimSpace(base)
		date = options.Now().Format("2006-01-02")
		archive = uniqueBranchName(options, projectRoot, fmt.Sprintf("sdlc_%s_state_%s", source, date))
		migration = uniqueBranchName(options, projectRoot, fmt.Sprintf("sdlc-v3-migration-%s", date))
		if err := options.RunCommand("git", []string{"branch", archive, "HEAD"}, projectRoot, nil, options.Output, options.ErrorOutput); err != nil {
			return fmt.Errorf("creating archive branch %s: %w", archive, err)
		}
		fmt.Fprintf(options.Output, "Archived the exact pre-migration state on %s.\n", archive)
		if yes, promptErr := promptYesNo(options, fmt.Sprintf("Push archive branch %s to origin? [y/N]: ", archive), false); promptErr != nil {
			return promptErr
		} else if yes {
			if err := options.RunCommand("git", []string{"push", "-u", "origin", archive}, projectRoot, nil, options.Output, options.ErrorOutput); err != nil {
				return fmt.Errorf("pushing archive branch %s: %w", archive, err)
			}
		}
		if err := options.RunCommand("git", []string{"switch", "-c", migration}, projectRoot, nil, options.Output, options.ErrorOutput); err != nil {
			return fmt.Errorf("creating migration branch %s: %w", migration, err)
		}
	}

	if resume == nil {
		if err := createInitializationWorkspace(sdlcRoot, projectRoot, workspace); err != nil {
			return fmt.Errorf("creating temporary initialization workspace %s: %w", workspace, err)
		}
	}
	generation := projectGeneration{values: values, explicit: explicit, source: source, base: base, archive: archive, migration: migration, date: date}
	if source == "v1" {
		if exists(filepath.Join(projectRoot, "docs", "ACs.md")) || exists(filepath.Join(projectRoot, "docs", "ACs.org")) {
			if err := prepareLegacyProject(options, projectRoot, values, runTicketMigration); err != nil {
				return err
			}
		}
	}
	if source == "v2" {
		legacyV2, archiveErr := archiveSpecKit(projectRoot)
		if archiveErr != nil {
			return archiveErr
		}
		generation.legacyV2 = legacyV2
		if err := ensureNoSpecKit(projectRoot); err != nil {
			return err
		}
	}
	if err := writeWorkLedger(sdlcRoot, projectRoot, generation, options.Now()); err != nil {
		return err
	}
	if exists(filepath.Join(projectRoot, "docs", "ACs.org")) {
		if err := repairLegacyLedgerIfRequired(options, sdlcRoot, projectRoot, values); err != nil {
			return err
		}
		merged, mergeErr := MergeLegacyAcceptanceCriteria(projectRoot)
		if mergeErr != nil {
			return fmt.Errorf("merging legacy acceptance criteria into docs/work.org: %w", mergeErr)
		}
		fmt.Fprintf(options.Output, "Merged %d legacy acceptance criteria into docs/work.org.\n", merged.AcceptanceCriteria)
	}
	if source == "v1" {
		if err := importLegacyWork(projectRoot); err != nil {
			return err
		}
	}
	if err := resolvePostMigrationConfiguration(options, schema, technologies, global, legacy, sdlcRoot, projectRoot, workspace, values, explicit, &generation); err != nil {
		return err
	}
	if source == "v2" {
		if err := writeWorkLedger(sdlcRoot, projectRoot, generation, options.Now()); err != nil {
			return err
		}
		if err := stripArchivedSpecStatuses(filepath.Join(projectRoot, "docs", "archive", "sdlc-v2"), generation.legacyV2); err != nil {
			return err
		}
	}
	if err := writeProjectProfile(projectRoot, schema, generation); err != nil {
		return err
	}
	if len(legacy) != 0 {
		if err := removeMigratedEnvironment(sdlcRoot, projectRoot, schema); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(workspace); err != nil {
		return fmt.Errorf("removing temporary initialization workspace %s: %w", workspace, err)
	}
	if err := options.RunCommand("git", []string{"add", "-A"}, projectRoot, nil, options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("staging migration: %w", err)
	}
	if err := commitMigration(options, projectRoot); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}
	fmt.Fprintf(options.Output, "Initialized SDLC v3 on %s.\n", migration)
	merge, err := promptYesNo(options, fmt.Sprintf("Merge %s into %s now? [y/N]: ", migration, base), false)
	if err != nil {
		return err
	}
	if merge {
		if err := options.RunCommand("git", []string{"switch", base}, projectRoot, nil, options.Output, options.ErrorOutput); err != nil {
			return fmt.Errorf("switching to %s: %w", base, err)
		}
		if err := options.RunCommand("git", []string{"merge", "--no-ff", migration}, projectRoot, nil, options.Output, options.ErrorOutput); err != nil {
			return fmt.Errorf("merging %s: %w", migration, err)
		}
	}
	return nil
}

func createInitializationWorkspace(sdlcRoot, projectRoot, workspace string) error {
	directory := filepath.Join(projectRoot, ".sdlc")
	if info, err := os.Lstat(directory); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return errors.New(".sdlc must be a project directory, not a file or symbolic link")
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(directory, 0o755); err != nil {
			return err
		}
	} else {
		return err
	}
	if err := ensureInitializationIgnore(sdlcRoot, directory); err != nil {
		return err
	}
	return os.Mkdir(workspace, 0o700)
}

func ensureInitializationIgnore(sdlcRoot, directory string) error {
	templatePath := filepath.Join(sdlcRoot, "templates", "v3", "project.gitignore")
	templateContents, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("reading project ignore template: %w", err)
	}
	destination := filepath.Join(directory, ".gitignore")
	info, err := os.Lstat(destination)
	if errors.Is(err, os.ErrNotExist) {
		// #nosec G306 -- this tracked project configuration is intentionally readable.
		return os.WriteFile(destination, templateContents, 0o644)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New(".sdlc/.gitignore must be a regular file, not a symbolic link")
	}
	existing, err := os.ReadFile(destination)
	if err != nil {
		return err
	}
	updated := string(existing)
	for _, required := range strings.Split(strings.TrimSpace(string(templateContents)), "\n") {
		if required == "" || containsExactLine(updated, required) {
			continue
		}
		if updated != "" && !strings.HasSuffix(updated, "\n") {
			updated += "\n"
		}
		updated += required + "\n"
	}
	if updated == string(existing) {
		return nil
	}
	// #nosec G306 -- this tracked project configuration is intentionally readable.
	return os.WriteFile(destination, []byte(updated), 0o644)
}

func containsExactLine(contents, wanted string) bool {
	for _, line := range strings.Split(contents, "\n") {
		if line == wanted {
			return true
		}
	}
	return false
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
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.RunCommand == nil {
		options.RunCommand = runCommand
	}
	if options.inputReader == nil {
		options.inputReader = bufio.NewReader(options.Input)
	}
	return options
}

func runCommand(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
	command := exec.Command(name, arguments...)
	command.Dir = directory
	command.Stdin = input
	command.Stdout = output
	command.Stderr = errorOutput
	return command.Run()
}

func commandOutput(options Options, directory, name string, arguments ...string) (string, error) {
	var output bytes.Buffer
	err := options.RunCommand(name, arguments, directory, nil, &output, options.ErrorOutput)
	return output.String(), err
}

func commitMigration(options Options, projectRoot string) error {
	arguments := []string{"commit", "-m", "chore: initialize lean SDLC v3"}
	if os.Getenv("VERBOSE") == "1" {
		return options.RunCommand("git", arguments, projectRoot, nil, options.Output, options.ErrorOutput)
	}

	var commitOutput bytes.Buffer
	var commitError bytes.Buffer
	if err := options.RunCommand("git", arguments, projectRoot, nil, &commitOutput, &commitError); err != nil {
		_, _ = io.Copy(options.Output, &commitOutput)
		_, _ = io.Copy(options.ErrorOutput, &commitError)
		return err
	}

	summary, err := commandOutput(options, projectRoot, "git", "show", "--shortstat", "--format=%h%x20%s", "HEAD")
	if err != nil || strings.TrimSpace(summary) == "" {
		fmt.Fprintln(options.Output, "Migration changes committed.")
		return nil
	}
	lines := make([]string, 0, 2)
	for _, line := range strings.Split(summary, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	fmt.Fprintf(options.Output, "Migration commit: %s.\n", strings.Join(lines, " | "))
	return nil
}

func ensureGitRepository(options Options, projectRoot string) error {
	output, err := commandOutput(options, projectRoot, "git", "rev-parse", "--is-inside-work-tree")
	if err == nil && strings.TrimSpace(output) == "true" {
		return nil
	}
	if err := options.RunCommand("git", []string{"init"}, projectRoot, nil, options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("initializing Git repository: %w", err)
	}
	fmt.Fprintln(options.Output, "Initialized a local Git repository for the project.")
	return nil
}

func ensureInitialCommit(options Options, projectRoot string) error {
	if err := options.RunCommand("git", []string{"rev-parse", "--verify", "HEAD"}, projectRoot, nil, io.Discard, io.Discard); err == nil {
		return nil
	}
	if err := options.RunCommand("git", []string{"commit", "--allow-empty", "-m", "chore: establish pre-SDLC state"}, projectRoot, nil, options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("creating initial recoverability commit: %w", err)
	}
	return nil
}

func requireCleanWorktree(options Options, projectRoot string) error {
	output, err := commandOutput(options, projectRoot, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("checking Git worktree: %w", err)
	}
	if strings.TrimSpace(output) != "" {
		return errors.New("project worktree is not clean; checkpoint or separate existing changes before migration")
	}
	return nil
}

// DetectSource classifies the project state without changing it.
func DetectSource(projectRoot string) (string, error) {
	if exists(filepath.Join(projectRoot, ".specify")) {
		return "v2", nil
	}
	for _, relative := range []string{"docs/ACs.md", "docs/ACs.org", "docs/ticket-migration.org"} {
		if exists(filepath.Join(projectRoot, relative)) {
			return "v1", nil
		}
	}
	return "new", nil
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// DiscoverTechnologies derives the selection from the deployed standards.
func DiscoverTechnologies(directory string) ([]Technology, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("reading technology standards: %w", err)
	}
	var technologies []Technology
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			continue
		}
		technologies = append(technologies, Technology{
			Name: strings.ToUpper(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))),
			Path: filepath.Join(directory, entry.Name()),
		})
	}
	sort.Slice(technologies, func(left, right int) bool { return technologies[left].Name < technologies[right].Name })
	return technologies, nil
}

func readYAMLConfig(path string) (map[string]any, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading global SDLC configuration %q: %w", path, err)
	}
	if len(bytes.TrimSpace(contents)) == 0 {
		return map[string]any{}, nil
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(false)
	values := map[string]any{}
	if err := decoder.Decode(&values); err != nil {
		return nil, fmt.Errorf("parsing global SDLC configuration %q: %w", path, err)
	}
	if rawVersion, present := values["version"]; present {
		version, ok := rawVersion.(int)
		if !ok || version != 3 {
			return nil, fmt.Errorf("global SDLC configuration %q must declare version: 3", path)
		}
	}
	return values, nil
}

func loadLegacyEnvironment(sdlcRoot, projectRoot string, schema ConfigSchema) (map[string]string, error) {
	path := filepath.Join(projectRoot, ".env")
	if !exists(path) {
		return map[string]string{}, nil
	}
	wrapper := filepath.Join(sdlcRoot, "libexec", "load-sdlc-env.sh")
	return configenv.Load(wrapper, path, schema.ManagedEnvironmentKeys())
}

func removeMigratedEnvironment(sdlcRoot, projectRoot string, schema ConfigSchema) error {
	path := filepath.Join(projectRoot, ".env")
	if !exists(path) {
		return nil
	}
	wrapper := filepath.Join(sdlcRoot, "libexec", "load-sdlc-env.sh")
	if err := configenv.RemoveKeys(wrapper, path, schema.ManagedEnvironmentKeys()); err != nil {
		return fmt.Errorf("retiring migrated SDLC values from project .env: %w", err)
	}
	return nil
}

func resolveConfiguration(options Options, schema ConfigSchema, technologies []Technology, global map[string]any, legacy map[string]string, technologyAssessment *technologyAssessment) (map[string]string, map[string]bool, error) {
	reader := options.inputReader
	values := map[string]string{}
	explicit := map[string]bool{}
	for _, field := range schema.Fields {
		if field.Phase == "post-migration" {
			continue
		}
		if !fieldApplies(field, values) {
			continue
		}
		value, isExplicit, valueSource := initialValue(field, options.Overrides, legacy, global)
		if field.Key == "SDLC_TECHNOLOGIES" && !isExplicit && value == "" && technologyAssessment != nil {
			value = technologyAssessment.selection()
			valueSource = "assessment"
			renderTechnologyAssessment(options.Output, *technologyAssessment)
		}
		prompted := false
		if field.Prompt != "" && !isExplicit && (valueSource != "global" || options.OverrideGlobalConfig) {
			selected, changed, err := promptField(reader, options.Output, field, value, valueSource, technologies)
			if err != nil {
				return nil, nil, err
			}
			value, isExplicit = selected, changed
			prompted = true
		}
		if err := field.ValidateValue(value, technologies); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", field.Key, err)
		}
		if prompted {
			renderResolvedSelection(options.Output, field, value)
		}
		if value != "" {
			values[field.Key] = value
		}
		if isExplicit || !field.AllowGlobal {
			explicit[field.Key] = true
		}
	}
	role := values["SDLC_INFRA_ROLE"]
	if role != "" && role != "none" {
		if values["SDLC_INFRA_OWNER"] == "" || values["SDLC_INFRA_CONTRACT"] == "" {
			return nil, nil, errors.New("infrastructure owner and contract are required for consumer or provider roles")
		}
	}
	return values, explicit, nil
}

func normalizeLegacyConfiguration(legacy map[string]string, diagnostics io.Writer) map[string]string {
	normalized := make(map[string]string, len(legacy))
	for key, value := range legacy {
		normalized[key] = value
	}
	for _, prefix := range []string{"SDLC_SPEC", "SDLC_BUILD", "SDLC_AUDIT"} {
		providerKey := prefix + "_PROVIDER"
		provider := strings.TrimSpace(normalized[providerKey])
		modelKey := prefix + "_MODEL"
		if prefix == "SDLC_SPEC" && provider == "" {
			providerKey = "SDLC_DELIVERY_PROVIDER"
			provider = strings.TrimSpace(normalized[providerKey])
			modelKey = "SDLC_DELIVERY_MODEL"
		}
		if provider == "" || provider == "openai" || provider == "openai-codex" {
			continue
		}
		delete(normalized, providerKey)
		delete(normalized, modelKey)
		fmt.Fprintf(diagnostics, "Warning: legacy %s provider %q is unsupported by SDLC v3; select an OpenAI model or inherit the global default.\n", strings.ToLower(strings.TrimPrefix(prefix, "SDLC_")), provider)
	}
	return normalized
}

func initialValue(field ConfigField, overrides, legacy map[string]string, global map[string]any) (string, bool, string) {
	if value := strings.TrimSpace(overrides[field.Key]); value != "" {
		return value, true, "cli"
	}
	if value := strings.TrimSpace(os.Getenv(field.Key)); value != "" {
		return value, true, "environment"
	}
	if value := strings.TrimSpace(legacy[field.Key]); value != "" {
		return normalizeLegacyValue(field, value), true, "legacy project environment"
	}
	for _, alias := range field.MigrationAliases {
		if value := strings.TrimSpace(legacy[alias]); value != "" {
			return normalizeLegacyValue(field, value), true, "legacy project environment"
		}
	}
	if field.AllowGlobal {
		if value, ok := yamlPathString(global, field.Path); ok {
			return value, false, "global"
		}
	}
	return field.Default, false, "schema"
}

func normalizeLegacyValue(field ConfigField, value string) string {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(field.Key, "_HARNESS") {
		return strings.ToLower(value)
	}
	if strings.HasSuffix(field.Key, "_PROVIDER") && value == "openai-codex" {
		return "openai"
	}
	return value
}

func promptField(reader *bufio.Reader, output io.Writer, field ConfigField, inherited, inheritedSource string, technologies []Technology) (string, bool, error) {
	choices := field.Choices
	if field.ChoicesFrom == "technologies" {
		for _, technology := range technologies {
			choices = append(choices, technology.Name)
		}
	}
	displayChoices := make([]promptChoice, 0, len(choices))
	selected := map[string]bool{}
	for _, value := range splitCSV(inherited) {
		selected[value] = true
	}
	for _, choice := range choices {
		displayChoices = append(displayChoices, promptChoice{Label: choice, Selected: selected[choice]})
	}
	renderPromptChoices(output, field.Prompt, displayChoices)
	if inherited != "" {
		label := "Default"
		action := "inherit"
		if inheritedSource == "global" {
			label = "Global SDLC default"
		} else if inheritedSource == "assessment" {
			label = "Recommended"
			action = "accept"
		}
		if field.Type == "multi-choice" {
			fmt.Fprintf(output, "%s: %s.\n", label, inherited)
		} else {
			fmt.Fprintf(output, "%s: %s. Press Enter to %s, or enter a project value.\n", label, inherited, action)
		}
	}
	if field.Type == "multi-choice" {
		return promptMultiChoice(reader, output, field, choices, selected)
	}
	fmt.Fprint(output, "Selection: ")
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", false, fmt.Errorf("reading %s: %w", field.Key, err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return inherited, false, nil
	}
	if field.Type == "choice" {
		for index, choice := range choices {
			if line == fmt.Sprint(index+1) {
				return choice, true, nil
			}
		}
	}
	return line, true, nil
}

func promptMultiChoice(reader *bufio.Reader, output io.Writer, field ConfigField, choices []string, selected map[string]bool) (string, bool, error) {
	changed := false
	for {
		fmt.Fprintln(output, "Press Enter to confirm, or enter numbers or names to toggle.")
		fmt.Fprint(output, "Selection: ")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", false, fmt.Errorf("reading %s: %w", field.Key, err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return selectedChoices(choices, selected), changed, nil
		}
		for _, item := range splitCSV(line) {
			choice, ok := resolveChoice(item, choices)
			if !ok {
				return "", false, fmt.Errorf("reading %s: unknown selection %q", field.Key, item)
			}
			selected[choice] = !selected[choice]
			changed = true
		}
		displayChoices := make([]promptChoice, 0, len(choices))
		for _, choice := range choices {
			displayChoices = append(displayChoices, promptChoice{Label: choice, Selected: selected[choice]})
		}
		renderPromptChoices(output, "Current selection:", displayChoices)
	}
}

func resolveChoice(value string, choices []string) (string, bool) {
	for index, choice := range choices {
		if value == choice || value == fmt.Sprint(index+1) {
			return choice, true
		}
	}
	return "", false
}

func selectedChoices(choices []string, selected map[string]bool) string {
	values := make([]string, 0, len(choices))
	for _, choice := range choices {
		if selected[choice] {
			values = append(values, choice)
		}
	}
	return strings.Join(values, ",")
}

func renderPromptChoices(output io.Writer, prompt string, choices []promptChoice) {
	fmt.Fprintf(output, "\n%s\n", prompt)
	for index, choice := range choices {
		mark := " "
		if choice.Selected {
			mark = "x"
		}
		fmt.Fprintf(output, "[%s] %d. %s\n", mark, index+1, choice.Label)
	}
}

func renderResolvedSelection(output io.Writer, field ConfigField, value string) {
	if value == "" {
		fmt.Fprintln(output, "Selected: [none]")
		return
	}
	if field.Type != "choice" && field.Type != "multi-choice" {
		fmt.Fprintf(output, "Selected: %s\n", value)
		return
	}
	selected := splitCSV(value)
	marked := make([]string, 0, len(selected))
	for _, item := range selected {
		marked = append(marked, "[x] "+item)
	}
	fmt.Fprintf(output, "Selected: %s\n", strings.Join(marked, ", "))
}

func yamlPathString(root map[string]any, path string) (string, bool) {
	var current any = root
	for _, part := range strings.Split(path, ".") {
		mapping, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		current, ok = mapping[part]
		if !ok {
			return "", false
		}
	}
	switch value := current.(type) {
	case string:
		return value, true
	case bool:
		return fmt.Sprint(value), true
	case []any:
		parts := make([]string, 0, len(value))
		for _, item := range value {
			parts = append(parts, fmt.Sprint(item))
		}
		return strings.Join(parts, ","), true
	default:
		return fmt.Sprint(value), true
	}
}

func promptYesNo(options Options, prompt string, defaultValue bool) (bool, error) {
	fmt.Fprint(options.Output, prompt)
	line, err := options.inputReader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		return defaultValue, nil
	default:
		return false, fmt.Errorf("expected yes or no, got %q", strings.TrimSpace(line))
	}
}

func uniqueBranchName(options Options, projectRoot, base string) string {
	name := base
	for suffix := 2; branchExists(options, projectRoot, name); suffix++ {
		name = fmt.Sprintf("%s-%d", base, suffix)
	}
	return name
}

func branchExists(options Options, projectRoot, name string) bool {
	err := options.RunCommand("git", []string{"show-ref", "--verify", "--quiet", "refs/heads/" + name}, projectRoot, nil, io.Discard, io.Discard)
	return err == nil
}

func interruptedMigration(options Options, projectRoot, workspace string) (*projectGeneration, error) {
	if !exists(workspace) {
		return nil, nil
	}
	current, err := commandOutput(options, projectRoot, "git", "branch", "--show-current")
	if err != nil {
		return nil, fmt.Errorf("reading interrupted migration branch: %w", err)
	}
	current = strings.TrimSpace(current)
	const prefix = "sdlc-v3-migration-"
	if !strings.HasPrefix(current, prefix) || len(current) < len(prefix)+10 {
		return nil, fmt.Errorf("cannot safely resume initialization from %s on branch %q", workspace, current)
	}
	date := current[len(prefix) : len(prefix)+10]
	branches, err := commandOutput(options, projectRoot, "git", "for-each-ref", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return nil, fmt.Errorf("listing archive branches for interrupted initialization: %w", err)
	}
	var archives []string
	for _, branch := range strings.Fields(branches) {
		if strings.HasPrefix(branch, "sdlc_") && strings.Contains(branch, "_state_"+date) {
			archives = append(archives, branch)
		}
	}
	if len(archives) != 1 {
		return nil, fmt.Errorf("cannot safely resume initialization from %s: expected one archive branch for %s, found %d", workspace, date, len(archives))
	}
	archive := archives[0]
	identity := strings.TrimPrefix(archive, "sdlc_")
	separator := strings.Index(identity, "_state_")
	if separator < 1 {
		return nil, fmt.Errorf("cannot safely derive the migration source from %s", archive)
	}
	base := ""
	for _, candidate := range []string{"master", "main"} {
		if !branchExists(options, projectRoot, candidate) {
			continue
		}
		archiveRevision, archiveErr := commandOutput(options, projectRoot, "git", "rev-parse", archive)
		baseRevision, baseErr := commandOutput(options, projectRoot, "git", "rev-parse", candidate)
		if archiveErr == nil && baseErr == nil && strings.TrimSpace(archiveRevision) == strings.TrimSpace(baseRevision) {
			base = candidate
			break
		}
	}
	if base == "" {
		return nil, fmt.Errorf("cannot safely derive the primary branch from %s", archive)
	}
	return &projectGeneration{
		source: identity[:separator], base: base, archive: archive, migration: current, date: date,
	}, nil
}

func chooseTicketMigration(options Options, projectRoot string) (bool, error) {
	if !exists(filepath.Join(projectRoot, "docs", "ACs.md")) || exists(filepath.Join(projectRoot, "docs", "ticket-migration.org")) {
		return false, nil
	}
	output, err := commandOutput(options, projectRoot, "gh", "issue", "list", "--state", "all", "--limit", "1", "--json", "number")
	if err != nil {
		remote, remoteErr := commandOutput(options, projectRoot, "git", "remote", "get-url", "origin")
		if remoteErr == nil && strings.Contains(strings.ToLower(remote), "github.com") {
			return false, errors.New("could not inspect GitHub issues for this GitHub-backed v1 project; authenticate gh or migrate the tickets before initialization")
		}
		fmt.Fprintln(options.ErrorOutput, "Warning: no usable GitHub issue source was found; only the legacy AC ledger will be converted.")
		return false, nil
	}
	var issues []map[string]any
	if json.Unmarshal([]byte(output), &issues) != nil || len(issues) == 0 {
		return false, nil
	}
	yes, err := promptYesNo(options, "Legacy GitHub tickets were found. Run the SDLC v1 ticket migration before initialization? [y/N]: ", false)
	if err != nil {
		return false, err
	}
	return yes, nil
}

func prepareLegacyProject(options Options, projectRoot string, values map[string]string, runTicketMigration bool) error {
	if runTicketMigration {
		if err := runConfiguredMutationSkill(options, projectRoot, values, "migrate-legacy-acs-to-sdlc-v1", "Use $migrate-legacy-acs-to-sdlc-v1 to prepare this project for SDLC v3 initialization. Complete the authorized migration before returning."); err != nil {
			return fmt.Errorf("running legacy ticket migration: %w", err)
		}
	}
	if exists(filepath.Join(projectRoot, "docs", "ACs.md")) {
		if err := runConfiguredMutationSkill(options, projectRoot, values, "convert-migrated-acs-to-org", "Use $convert-migrated-acs-to-org to convert docs/ACs.md losslessly to the canonical docs/ACs.org before SDLC v3 initialization. Do not change requirements."); err != nil {
			return fmt.Errorf("converting the legacy AC ledger: %w", err)
		}
	}
	if !exists(filepath.Join(projectRoot, "docs", "ACs.org")) {
		return errors.New("legacy migration did not produce docs/ACs.org")
	}
	return nil
}

func runConfiguredMutationSkill(options Options, projectRoot string, values map[string]string, skill, prompt string) error {
	harnessName := strings.TrimSpace(values["SDLC_AUDIT_HARNESS"])
	if harnessName == "" {
		harnessName = "codex"
	}
	if harnessName != "codex" {
		fmt.Fprintf(options.Output, "Harness: %s\n", harnessName)
		fmt.Fprintf(options.Output, "Executable: %s\n", harnessName)
		fmt.Fprintf(options.Output, "Project: %s\n", projectRoot)
		fmt.Fprintf(options.Output, "Skill: $%s\n", skill)
		fmt.Fprintln(options.Output, "Scope: mutate only the named migration artefacts in this project")
		fmt.Fprintln(options.Output, "Expected artefacts: the files required by the named skill")
		fmt.Fprintln(options.Output, "Pending stage: project initialization")
		fmt.Fprintln(options.Output, "Recovery record: .sdlc/.init")
		fmt.Fprintln(options.Output, "Resume command: sdlc-init")
		return fmt.Errorf("%s has no proven confined project-write adapter; complete the handoff and rerun sdlc-init", harnessName)
	}
	arguments := []string{"exec"}
	if model := values["SDLC_AUDIT_MODEL"]; model != "" {
		arguments = append(arguments, "--model", model)
	}
	arguments = append(arguments, prompt)
	return options.RunCommand("codex", arguments, projectRoot, nil, options.Output, options.ErrorOutput)
}

func repairLegacyLedgerIfRequired(options Options, sdlcRoot, projectRoot string, values map[string]string) error {
	ledgerPath := filepath.Join(projectRoot, "docs", "ACs.org")
	ledger, err := readRegularFile(ledgerPath)
	if err != nil {
		return fmt.Errorf("reading legacy acceptance-criteria ledger: %w", err)
	}
	if _, _, validationErr := renderLegacyLedgerBlock(string(ledger)); validationErr == nil {
		return nil
	} else {
		promptPath := filepath.Join(sdlcRoot, "prompts", "normalize-legacy-acs.md")
		prompt, readErr := os.ReadFile(promptPath)
		if readErr != nil {
			return fmt.Errorf("reading legacy-ledger repair prompt: %w", readErr)
		}
		fmt.Fprintf(options.Output, "Legacy AC ledger is non-canonical; invoking one repair agent.\n")
		if harnessName := strings.TrimSpace(values["SDLC_AUDIT_HARNESS"]); harnessName != "" && harnessName != "codex" {
			return runConfiguredMutationSkill(options, projectRoot, values, "convert-migrated-acs-to-org", string(prompt))
		}
		prompt = append(prompt, []byte("\n\nCanonical template: "+filepath.Join(sdlcRoot, "templates", "migration", "ACs.org")+"\nInitial validation failure: "+validationErr.Error()+"\n")...)
		arguments := []string{"exec", "--ephemeral", "--sandbox", "workspace-write"}
		if model := strings.TrimSpace(values["SDLC_AUDIT_MODEL"]); model != "" {
			arguments = append(arguments, "--model", model)
		}
		arguments = append(arguments, "-")
		if err := options.RunCommand("codex", arguments, projectRoot, bytes.NewReader(prompt), options.Output, options.ErrorOutput); err != nil {
			return fmt.Errorf("normalizing docs/ACs.org with headless Codex: %w", err)
		}
	}
	repaired, err := readRegularFile(ledgerPath)
	if err != nil {
		return fmt.Errorf("reading repaired legacy acceptance-criteria ledger: %w", err)
	}
	if _, _, err := renderLegacyLedgerBlock(string(repaired)); err != nil {
		return fmt.Errorf("repair agent left docs/ACs.org non-canonical: %w", err)
	}
	return nil
}

func archiveSpecKit(projectRoot string) ([]string, error) {
	archiveRoot := filepath.Join(projectRoot, "docs", "archive", "sdlc-v2")
	var workItems []string
	for _, relative := range []string{".specify", "specs"} {
		source := filepath.Join(projectRoot, relative)
		if !exists(source) {
			continue
		}
		if relative == "specs" {
			entries, err := os.ReadDir(source)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				if entry.IsDir() {
					workItems = append(workItems, entry.Name())
				}
			}
		}
		destination := filepath.Join(archiveRoot, relative)
		if exists(destination) {
			return nil, fmt.Errorf("both active and archived Spec Kit paths exist: %s and %s", source, destination)
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return nil, err
		}
		if err := os.Rename(source, destination); err != nil {
			return nil, fmt.Errorf("moving Spec Kit state into the project archive: %w", err)
		}
	}
	for _, relative := range []string{".agents/skills", ".claude/commands", ".codex/prompts"} {
		path := filepath.Join(projectRoot, relative)
		if exists(path) {
			destination := filepath.Join(archiveRoot, "integrations", relative)
			if err := archiveSpecKitEntries(path, destination); err != nil {
				return nil, fmt.Errorf("archiving project-local Spec Kit integration %s: %w", relative, err)
			}
		}
	}
	entries, err := os.ReadDir(filepath.Join(archiveRoot, "specs"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading archived Spec Kit work: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			workItems = append(workItems, entry.Name())
		}
	}
	workItems = uniqueSortedStrings(workItems)
	return workItems, nil
}

func uniqueSortedStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func stripArchivedSpecStatuses(archiveRoot string, workItems []string) error {
	for _, item := range workItems {
		path := filepath.Join(archiveRoot, "specs", item, "spec.md")
		contents, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("reading archived Spec Kit specification %s: %w", item, err)
		}
		updated, changed := stripSpecStatusField(string(contents))
		if !changed {
			continue
		}
		if err := writeAtomic(path, []byte(updated)); err != nil {
			return fmt.Errorf("removing duplicated lifecycle status from archived Spec Kit specification %s: %w", item, err)
		}
	}
	return nil
}

func stripSpecStatusField(document string) (string, bool) {
	lines := strings.Split(document, "\n")
	for index, line := range lines {
		if index > 24 || strings.HasPrefix(line, "## ") {
			break
		}
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if !strings.HasPrefix(lower, "status:") && !strings.HasPrefix(lower, "**status**:") {
			continue
		}
		lines = append(lines[:index], lines[index+1:]...)
		if index < len(lines) && strings.TrimSpace(lines[index]) == "" && index > 0 && strings.TrimSpace(lines[index-1]) == "" {
			lines = append(lines[:index], lines[index+1:]...)
		}
		return strings.Join(lines, "\n"), true
	}
	return document, false
}

func archiveSpecKitEntries(directory, archiveDirectory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := strings.ToLower(entry.Name())
		if name != "speckit" && !strings.HasPrefix(name, "speckit-") && !strings.HasPrefix(name, "speckit.") {
			continue
		}
		if err := os.MkdirAll(archiveDirectory, 0o755); err != nil {
			return err
		}
		if err := os.Rename(filepath.Join(directory, entry.Name()), filepath.Join(archiveDirectory, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func ensureNoSpecKit(projectRoot string) error {
	for _, relative := range []string{".specify", "specs"} {
		if exists(filepath.Join(projectRoot, relative)) {
			return fmt.Errorf("active Spec Kit artefact remains after archival: %s", relative)
		}
	}
	return nil
}

func writeProjectProfile(projectRoot string, schema ConfigSchema, generation projectGeneration) error {
	root := map[string]any{"version": 3}
	setYAMLPath(root, "project.migration_source", generation.source)
	setYAMLPath(root, "project.primary_branch", generation.base)
	setYAMLPath(root, "migration.archive_branch", generation.archive)
	setYAMLPath(root, "migration.branch", generation.migration)
	setYAMLPath(root, "migration.date", generation.date)
	for _, field := range schema.Fields {
		value := generation.values[field.Key]
		if !generation.explicit[field.Key] {
			continue
		}
		if value == "" {
			if field.Type == "string-list" {
				setYAMLPath(root, field.Path, []string{})
			}
			continue
		}
		var stored any = value
		if field.Type == "boolean" {
			stored = value == "true"
		} else if field.Type == "multi-choice" || field.Type == "string-list" {
			stored = splitCSV(value)
		}
		setYAMLPath(root, field.Path, stored)
	}
	if requirements := splitCSV(generation.values["SDLC_REQUIREMENT_AUTHORITIES"]); len(requirements) != 0 {
		setYAMLPath(root, "authorities.requirements", requirements)
	}
	contents, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("rendering project profile: %w", err)
	}
	directory := filepath.Join(projectRoot, ".sdlc")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	// #nosec G306 -- this tracked, non-secret project profile is intentionally readable.
	return os.WriteFile(filepath.Join(directory, "project.yaml"), contents, 0o644)
}

func setYAMLPath(root map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := root
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	current[parts[len(parts)-1]] = value
}

func writeWorkLedger(sdlcRoot, projectRoot string, generation projectGeneration, now time.Time) error {
	templatePath := filepath.Join(sdlcRoot, "templates", "v3", "work.org")
	contents, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("reading work ledger template: %w", err)
	}
	replacements := map[string]string{
		"[source]":     generation.source,
		"[YYYY-MM-DD]": now.Format("2006-01-02"),
		"[Dated archive branch and optional origin link.]": generation.archive,
		"[Branch name.]":              generation.migration,
		"{{TICKET_ARCHIVE}}":          "Not applicable.",
		"{{HISTORICAL_REQUIREMENTS}}": "Not applicable.",
		"{{V2_ARCHIVE}}":              "Not applicable.",
	}
	if generation.source == "v1" {
		replacements["{{TICKET_ARCHIVE}}"] = "[[file:archive/migrated-tickets/][Archived ticket corpus]]."
		replacements["{{HISTORICAL_REQUIREMENTS}}"] = "[[#legacy-acceptance-criteria][Legacy Acceptance Criteria (SDLC v1)]]."
	}
	if generation.source == "v2" {
		replacements["{{V2_ARCHIVE}}"] = "[[file:archive/sdlc-v2/][Archived Spec Kit work]]."
	}
	text := string(contents)
	for before, after := range replacements {
		text = strings.ReplaceAll(text, before, after)
	}
	if !exists(filepath.Join(projectRoot, "docs", "ACs.org")) {
		text = removeEmptyLegacyLedgerSection(text)
	}
	directory := filepath.Join(projectRoot, "docs")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	workPath := filepath.Join(directory, "work.org")
	if exists(workPath) {
		existing, err := readRegularFile(workPath)
		if err != nil {
			return fmt.Errorf("reading existing work ledger: %w", err)
		}
		text, err = mergeWorkLedgerScaffold(string(existing), text, generation.migration)
		if err != nil {
			return err
		}
	}
	if len(generation.migratedV2) != 0 {
		nextWork := highestHistoricalWorkNumber(projectRoot) + 1
		var entries []string
		for _, item := range generation.migratedV2 {
			if strings.Contains(text, ":SOURCE: "+item.Path+"\n") {
				continue
			}
			identifier := fmt.Sprintf("W%03d", nextWork)
			nextWork++
			legacyDirectory := filepath.Base(filepath.Dir(filepath.FromSlash(item.Path)))
			legacyIdentifier, _ := v2WorkIdentity(legacyDirectory)
			state := migratedWorkState(item.Disposition)
			descriptor := strings.NewReplacer("[", "(", "]", ")").Replace(item.Descriptor)
			var entry strings.Builder
			fmt.Fprintf(&entry, "** %s %s - %s :feature:migration:\n", state, identifier, descriptor)
			entry.WriteString(":PROPERTIES:\n")
			fmt.Fprintf(&entry, ":CUSTOM_ID: w-%s\n", strings.ToLower(strings.TrimPrefix(identifier, "W")))
			fmt.Fprintf(&entry, ":LEGACY_ID: %s\n", legacyIdentifier)
			fmt.Fprintf(&entry, ":PRIORITY: %s\n", item.Priority)
			fmt.Fprintf(&entry, ":SOURCE: %s\n", item.Path)
			fmt.Fprintf(&entry, ":CREATED: %s\n", item.Created)
			fmt.Fprintf(&entry, ":MIGRATION_DISPOSITION: %s\n", item.Disposition)
			entry.WriteString(":END:\n\n")
			linkPath := strings.TrimPrefix(item.Path, "docs/")
			fmt.Fprintf(&entry, "- *Specification:* [[file:%s][%s]]\n", linkPath, descriptor)
			fmt.Fprintf(&entry, "- *Migration evidence:* %s\n", item.Evidence)
			entries = append(entries, entry.String())
		}
		if len(entries) != 0 {
			text, err = appendToOrgTopLevelSection(text, "Work items", strings.Join(entries, "\n"))
			if err != nil {
				return err
			}
		}
	}
	// #nosec G306 -- this tracked project document is intentionally readable.
	return os.WriteFile(workPath, []byte(text), 0o644)
}

func mergeWorkLedgerScaffold(existing, rendered, migrationBranch string) (string, error) {
	result := mergeWorkLedgerPreamble(existing, rendered)
	if legacy := legacyLedgerSection(result); legacy != "" {
		normalized, _, err := normalizeLegacyLedgerBlock(legacy)
		if err != nil {
			return "", err
		}
		result = strings.Replace(result, legacy, normalized, 1)
	}
	canonical := splitOrgTopLevelSections(strings.Split(rendered, "\n"))
	if len(canonical) == 0 {
		return "", errors.New("canonical work ledger template contains no level-one sections")
	}
	for _, section := range canonical {
		if section.title == "Migration record" {
			continue
		}
		if strings.HasPrefix(section.title, legacyLedgerTitle) && legacyLedgerSection(result) != "" {
			continue
		}
		marker := "* " + section.title + "\n"
		if strings.Contains(result, marker) {
			continue
		}
		addition := strings.Join(section.lines, "\n") + "\n\n"
		migrationMarker := "* Migration record\n"
		if strings.Contains(result, migrationMarker) {
			result = strings.Replace(result, migrationMarker, addition+migrationMarker, 1)
		} else {
			result = strings.TrimRight(result, "\n") + "\n\n" + addition
		}
	}
	var err error
	result, err = consolidateStateSections(result)
	if err != nil {
		return "", err
	}

	var migration orgSection
	for _, section := range canonical {
		if section.title == "Migration record" {
			migration = section
			break
		}
	}
	if migration.title == "" {
		return "", errors.New("canonical work ledger template lacks the Migration record section")
	}
	migrationMarker := "* Migration record\n"
	if !strings.Contains(result, migrationMarker) {
		return strings.TrimRight(result, "\n") + "\n\n" + strings.Join(migration.lines, "\n") + "\n", nil
	}
	if migrationBranch != "" && !strings.Contains(result, "- *Migration branch:* "+migrationBranch+"\n") {
		body := strings.Join(trimBlankEdges(migration.lines[1:]), "\n")
		result = strings.Replace(result, migrationMarker, migrationMarker+"\n"+body+"\n", 1)
	}
	return result, nil
}

func migratedWorkState(disposition string) string {
	switch disposition {
	case "delivered":
		return "DONE"
	case "approved-undelivered":
		return "TODO"
	case "abandoned":
		return "ABANDONED"
	default:
		return "REVIEW"
	}
}

func v2WorkIdentity(directory string) (string, string) {
	parts := strings.SplitN(directory, "-", 2)
	identifier := strings.ToUpper(parts[0])
	descriptor := directory
	if len(parts) == 2 {
		descriptor = strings.ReplaceAll(parts[1], "-", " ")
	}
	return identifier, descriptor
}

func highestHistoricalWorkNumber(projectRoot string) int {
	highest := 0
	for _, relative := range []string{"docs/ACs.org", "docs/ACs.md", "docs/ticket-migration.org", "docs/work.org"} {
		contents, err := os.ReadFile(filepath.Join(projectRoot, relative))
		if err != nil {
			continue
		}
		for _, match := range historicalIdentifierPattern.FindAllStringSubmatch(withoutOrgExamples(string(contents)), -1) {
			value := match[1]
			if value == "" {
				value = match[2]
			}
			number, err := strconv.Atoi(value)
			if err == nil && number > highest {
				highest = number
			}
		}
	}
	return highest
}

func withoutOrgExamples(document string) string {
	var output []string
	ignored := ""
	for _, line := range strings.Split(document, "\n") {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if ignored == "" && (trimmed == "#+begin_example" || trimmed == "#+begin_comment") {
			ignored = strings.TrimPrefix(trimmed, "#+begin_")
			continue
		}
		if ignored != "" {
			if trimmed == "#+end_"+ignored {
				ignored = ""
			}
			continue
		}
		output = append(output, line)
	}
	return strings.Join(output, "\n")
}

func importLegacyWork(projectRoot string) error {
	migrationPath := filepath.Join(projectRoot, "docs", "ticket-migration.org")
	if !exists(migrationPath) {
		return nil
	}
	source, err := os.ReadFile(migrationPath)
	if err != nil {
		return err
	}
	workPath := filepath.Join(projectRoot, "docs", "work.org")
	work, err := os.ReadFile(workPath)
	if err != nil {
		return err
	}
	type section struct {
		source, state, tags string
	}
	sections := []section{
		{"Open defects at migration", "TODO", "defect:legacy"},
		{"Defined but undelivered features", "TODO", "feature:legacy"},
		{"Requires human review", "REVIEW", "legacy:migration"},
	}
	result := string(work)
	var entries []string
	for _, item := range sections {
		body := orgTopLevelBody(string(source), item.source)
		converted, convertErr := convertLegacyTicketHeadings(body, item.state, item.tags)
		if convertErr != nil {
			return convertErr
		}
		if strings.TrimSpace(converted) == "" {
			continue
		}
		for _, entry := range splitOrgLevelTwoEntries(converted) {
			source := workEntryProperty(entry, "SOURCE")
			if source != "" && strings.Contains(result, ":SOURCE: "+source+"\n") {
				continue
			}
			entries = append(entries, entry)
		}
	}
	if len(entries) != 0 {
		result, err = appendToOrgTopLevelSection(result, "Work items", strings.Join(entries, "\n\n"))
		if err != nil {
			return err
		}
	}
	// #nosec G306 -- this tracked project document is intentionally readable.
	return os.WriteFile(workPath, []byte(result), 0o644)
}

func splitOrgLevelTwoEntries(document string) []string {
	lines := strings.Split(strings.TrimSpace(document), "\n")
	var entries []string
	start := -1
	for index, line := range lines {
		if !strings.HasPrefix(line, "** ") {
			continue
		}
		if start >= 0 {
			entries = append(entries, strings.TrimSpace(strings.Join(lines[start:index], "\n")))
		}
		start = index
	}
	if start >= 0 {
		entries = append(entries, strings.TrimSpace(strings.Join(lines[start:], "\n")))
	}
	return entries
}

func workEntryProperty(entry, name string) string {
	prefix := ":" + name + ":"
	for _, line := range strings.Split(entry, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func orgTopLevelBody(document, title string) string {
	lines := strings.Split(document, "\n")
	start := -1
	for index, line := range lines {
		if line == "* "+title {
			start = index + 1
			continue
		}
		if start >= 0 && strings.HasPrefix(line, "* ") {
			return strings.Join(lines[start:index], "\n")
		}
	}
	if start >= 0 {
		return strings.Join(lines[start:], "\n")
	}
	return ""
}

func convertLegacyTicketHeadings(body, state, tags string) (string, error) {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	var output []string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "- None." || strings.TrimSpace(line) == "=- None.=" {
			continue
		}
		if strings.HasPrefix(line, "** [[") {
			number, descriptor, source, err := parseLegacyTicketHeading(line)
			if err != nil {
				return "", err
			}
			output = append(output,
				fmt.Sprintf("** %s W%03d - %s :%s:", state, number, descriptor, tags),
				":PROPERTIES:",
				fmt.Sprintf(":CUSTOM_ID: w-%03d", number),
				":PRIORITY: unassigned",
				fmt.Sprintf(":SOURCE: %s", source),
				":CREATED: unknown",
				":END:",
			)
			continue
		}
		if strings.HasPrefix(line, "***") {
			line = "*" + line
		}
		output = append(output, line)
	}
	return strings.TrimSpace(strings.Join(output, "\n")), nil
}

func parseLegacyTicketHeading(line string) (int, string, string, error) {
	start := strings.Index(line, "[[")
	middle := strings.Index(line, "][#")
	end := strings.LastIndex(line, "]]")
	if start < 0 || middle < 0 || end < middle {
		return 0, "", "", fmt.Errorf("cannot parse legacy ticket heading %q", line)
	}
	target := line[start+2 : middle]
	label := line[middle+2 : end]
	parts := strings.SplitN(strings.TrimPrefix(label, "#"), " - ", 2)
	if len(parts) != 2 {
		return 0, "", "", fmt.Errorf("legacy ticket heading lacks descriptor %q", line)
	}
	number, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", "", fmt.Errorf("legacy ticket heading has invalid number %q", line)
	}
	return number, parts[1], "[[" + target + "][" + label + "]]", nil
}
