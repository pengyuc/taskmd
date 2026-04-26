package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/template"
)

var (
	projectInitForce       bool
	projectInitStdout      bool
	projectInitClaude      bool
	projectInitGemini      bool
	projectInitCodex       bool
	projectInitNoSpec      bool
	projectInitNoAgent     bool
	projectInitNoTemplates bool
	projectInitTaskDir     string
	projectInitIDStrategy  string
	projectInitIDPrefix    string
)

// projectInitRoot is the project root directory. Defaults to ".".
// Tests override this to t.TempDir() for parallel safety.
var projectInitRoot = "."

// projectInitIsTTY checks whether stdin is a terminal.
// Tests override this to return false.
var projectInitIsTTY = func() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
}

const configFilename = ".taskmd.yaml"

var projectInitCmd = &cobra.Command{
	Use:        "init",
	SuggestFor: []string{"setup", "create", "new"},
	Short:      "Initialize a taskmd project with config, task directory, and agent files",
	Long: `Initialize sets up a complete taskmd project in the current directory.

Creates a task directory, .taskmd.yaml config, agent configuration files, and
the taskmd specification document. When run interactively, prompts for any
values not provided via flags.

If a file already exists and --force is not set, it is skipped with a warning.

Examples:
  taskmd init                        # Interactive setup (prompts for missing info)
  taskmd init --task-dir ./tasks     # Set task directory, prompt for agents
  taskmd init --claude               # Claude agent, prompt for task directory
  taskmd init --task-dir ./tasks --claude  # Fully non-interactive
  taskmd init --claude --gemini      # Multiple agents
  taskmd init --no-spec              # Skip TASKMD_SPEC.md
  taskmd init --no-agent             # Skip agent configs
  taskmd init --no-templates         # Skip task templates
  taskmd init --force                # Overwrite existing files
  taskmd init --stdout               # Print all content to stdout`,
	Args: cobra.NoArgs,
	RunE: runProjectInit,
}

func init() {
	rootCmd.AddCommand(projectInitCmd)

	projectInitCmd.Flags().BoolVar(&projectInitForce, "force", false, "overwrite existing files")
	projectInitCmd.Flags().BoolVar(&projectInitStdout, "stdout", false, "print all content to stdout instead of writing files")
	projectInitCmd.Flags().BoolVar(&projectInitClaude, "claude", false, "initialize for Claude Code")
	projectInitCmd.Flags().BoolVar(&projectInitGemini, "gemini", false, "initialize for Gemini")
	projectInitCmd.Flags().BoolVar(&projectInitCodex, "codex", false, "initialize for Codex")
	projectInitCmd.Flags().BoolVar(&projectInitNoSpec, "no-spec", false, "skip generating TASKMD_SPEC.md")
	projectInitCmd.Flags().BoolVar(&projectInitNoAgent, "no-agent", false, "skip generating agent configuration files")
	projectInitCmd.Flags().BoolVar(&projectInitNoTemplates, "no-templates", false, "skip copying built-in task templates")
	projectInitCmd.Flags().StringVar(&projectInitTaskDir, "task-dir", "./tasks", "task directory path to create")
	projectInitCmd.Flags().StringVar(&projectInitIDStrategy, "id-strategy", "", "ID generation strategy: sequential, prefixed, random, ulid")
	projectInitCmd.Flags().StringVar(&projectInitIDPrefix, "id-prefix", "", "prefix for prefixed ID strategy")
}

// fileToWrite represents a file that the init command will create.
type fileToWrite struct {
	filename string
	content  []byte
}

func runProjectInit(cmd *cobra.Command, _ []string) error {
	if projectInitNoSpec && projectInitNoAgent && projectInitNoTemplates {
		return fmt.Errorf("--no-spec, --no-agent, and --no-templates cannot all be set (nothing to do)")
	}

	root := projectInitRoot
	isTTY := projectInitIsTTY()
	quiet := GetGlobalFlags().Quiet

	// Resolve task directory: flag > prompt > default
	taskDirPath, err := resolveInitTaskDir(cmd, isTTY)
	if err != nil {
		return err
	}

	// Resolve agents: flags > prompt > default (Claude)
	resolveInitAgents(isTTY)

	// Resolve templates: flags > prompt > default (yes)
	resolveInitTemplates(isTTY)

	// Resolve ID strategy: flag > prompt > default (sequential)
	idStrategy, err := resolveInitIDStrategy(cmd, isTTY)
	if err != nil {
		return err
	}

	// Collect files split by destination
	rootFiles, taskDirFiles := collectInitFiles(idStrategy)

	// --stdout mode: print everything and exit
	if projectInitStdout {
		allFiles := append(rootFiles, taskDirFiles...)
		return printFilesToStdout(allFiles)
	}

	taskDirAbs := taskDirPath
	if !filepath.IsAbs(taskDirAbs) {
		taskDirAbs = filepath.Join(root, taskDirPath)
	}

	// Show confirmation before writing
	if isTTY && !quiet {
		if !confirmInitPlan(root, taskDirAbs, rootFiles, taskDirFiles) {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return nil
		}
	}

	return writeProjectFiles(root, taskDirAbs, taskDirPath, idStrategy, rootFiles, taskDirFiles, quiet)
}

// writeProjectFiles creates directories, config, and all init files.
func writeProjectFiles(root, taskDirAbs, taskDirPath string, idStrategy idStrategyConfig, rootFiles, taskDirFiles []fileToWrite, quiet bool) error {
	var createdPaths []string

	dirCreated, err := ensureTaskDir(taskDirAbs, quiet)
	if err != nil {
		return err
	}
	if dirCreated {
		abs, _ := filepath.Abs(taskDirAbs)
		createdPaths = append(createdPaths, abs+"/")
	}

	configCreated, err := writeConfigFile(root, taskDirPath, idStrategy, quiet)
	if err != nil {
		return err
	}
	if configCreated {
		abs, _ := filepath.Abs(filepath.Join(root, configFilename))
		createdPaths = append(createdPaths, abs)
	}

	rootCreated, err := writeInitFiles(root, rootFiles, quiet)
	if err != nil {
		return err
	}
	createdPaths = append(createdPaths, rootCreated...)

	tdCreated, err := writeInitFiles(taskDirAbs, taskDirFiles, quiet)
	if err != nil {
		return err
	}
	createdPaths = append(createdPaths, tdCreated...)

	// Write built-in templates to .taskmd/templates/
	if !projectInitNoTemplates {
		tmplCreated, err := writeBuiltinTemplates(root, quiet)
		if err != nil {
			return err
		}
		createdPaths = append(createdPaths, tmplCreated...)
	}

	if !quiet {
		printInitSummary(createdPaths)
	}

	return nil
}

// writeBuiltinTemplates copies built-in task templates to .taskmd/templates/.
func writeBuiltinTemplates(root string, quiet bool) ([]string, error) {
	tmplDir := filepath.Join(root, ".taskmd", "templates")
	if err := os.MkdirAll(tmplDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	var files []fileToWrite
	for name, content := range template.BuiltinTemplates {
		files = append(files, fileToWrite{
			filename: name + ".md",
			content:  []byte(content),
		})
	}

	return writeInitFiles(tmplDir, files, quiet)
}

// resolveInitTaskDir returns the task directory path.
// If --task-dir was explicitly provided, uses that.
// If TTY, prompts the user. Otherwise uses the default.
func resolveInitTaskDir(cmd *cobra.Command, isTTY bool) (string, error) {
	if cmd.Flags().Changed("task-dir") {
		return projectInitTaskDir, nil
	}

	if isTTY {
		value := projectInitTaskDir // default for the prompt
		err := huh.NewInput().
			Title("Task directory").
			Value(&value).
			Run()
		if err != nil {
			return "", fmt.Errorf("prompt cancelled: %w", err)
		}
		return value, nil
	}

	return projectInitTaskDir, nil
}

// resolveInitAgents sets agent flags via prompt if none were explicitly set.
func resolveInitAgents(isTTY bool) {
	// If any agent flag is set, respect it
	if projectInitClaude || projectInitGemini || projectInitCodex || projectInitNoAgent {
		return
	}

	if isTTY {
		promptAgentSelection()
		return
	}

	// Non-TTY with no agent flags: default to Claude
	projectInitClaude = true
}

// promptAgentSelection shows a multi-select for agent configs.
func promptAgentSelection() {
	options := []huh.Option[string]{
		huh.NewOption("Claude Code", "claude").Selected(true),
		huh.NewOption("Gemini", "gemini"),
		huh.NewOption("Codex", "codex"),
	}

	var selected []string
	err := huh.NewMultiSelect[string]().
		Title("Which AI assistants do you use?").
		Options(options...).
		Value(&selected).
		Run()
	if err != nil || len(selected) == 0 {
		// Cancelled or nothing selected: default to Claude
		projectInitClaude = true
		return
	}

	for _, s := range selected {
		switch s {
		case "claude":
			projectInitClaude = true
		case "gemini":
			projectInitGemini = true
		case "codex":
			projectInitCodex = true
		}
	}
}

// resolveInitTemplates sets projectInitNoTemplates via prompt if not explicitly set.
func resolveInitTemplates(isTTY bool) {
	if projectInitNoTemplates {
		return
	}

	if !isTTY {
		return
	}

	include := true
	err := huh.NewConfirm().
		Title("Include task templates?").
		Description("Copies built-in templates (feature, bug, chore) to .taskmd/templates/").
		Value(&include).
		WithButtonAlignment(lipgloss.Left).
		Run()
	if err != nil {
		// Cancelled: default to including templates
		return
	}
	projectInitNoTemplates = !include
}

// confirmInitPlan shows what will be created and asks the user to confirm.
// Returns true if the user confirms, false if cancelled.
func confirmInitPlan(root, taskDirAbs string, rootFiles, taskDirFiles []fileToWrite) bool {
	fmt.Println("\nThe following will be created:")

	// Task directory
	abs, _ := filepath.Abs(taskDirAbs)
	fmt.Printf("  %s/\n", abs)

	// Config file
	configAbs, _ := filepath.Abs(filepath.Join(root, configFilename))
	fmt.Printf("  %s\n", configAbs)

	// Root files
	for _, f := range rootFiles {
		p, _ := filepath.Abs(filepath.Join(root, f.filename))
		fmt.Printf("  %s\n", p)
	}

	// Task dir files
	for _, f := range taskDirFiles {
		p, _ := filepath.Abs(filepath.Join(taskDirAbs, f.filename))
		fmt.Printf("  %s\n", p)
	}

	// Templates
	if !projectInitNoTemplates {
		tmplAbs, _ := filepath.Abs(filepath.Join(root, ".taskmd", "templates"))
		fmt.Printf("  %s/\n", tmplAbs)
	}

	fmt.Println()

	confirm := true
	err := huh.NewConfirm().
		Title("Proceed?").
		Value(&confirm).
		Run()
	if err != nil {
		return false
	}
	return confirm
}

// idStrategyConfig holds the resolved ID strategy and associated settings.
type idStrategyConfig struct {
	strategy string // sequential, prefixed, random, ulid
	prefix   string // only for prefixed
}

var validIDStrategies = []string{"sequential", "prefixed", "random", "ulid"}

// resolveInitIDStrategy returns the ID strategy config.
// If --id-strategy was provided, uses it. If TTY, prompts. Otherwise defaults to sequential.
func resolveInitIDStrategy(cmd *cobra.Command, isTTY bool) (idStrategyConfig, error) {
	cfg := idStrategyConfig{strategy: "sequential"}

	if cmd.Flags().Changed("id-strategy") {
		if !isValidIDStrategy(projectInitIDStrategy) {
			return cfg, fmt.Errorf("invalid --id-strategy %q: must be one of %s",
				projectInitIDStrategy, strings.Join(validIDStrategies, ", "))
		}
		cfg.strategy = projectInitIDStrategy
		if cfg.strategy == "prefixed" {
			prefix, err := resolveIDPrefix(cmd, isTTY)
			if err != nil {
				return cfg, err
			}
			cfg.prefix = prefix
		}
		return cfg, nil
	}

	if isTTY {
		strategy, err := promptIDStrategy()
		if err != nil {
			return cfg, err
		}
		cfg.strategy = strategy
		if cfg.strategy == "prefixed" {
			prefix, err := resolveIDPrefix(cmd, isTTY)
			if err != nil {
				return cfg, err
			}
			cfg.prefix = prefix
		}
		return cfg, nil
	}

	return cfg, nil
}

func isValidIDStrategy(s string) bool {
	for _, v := range validIDStrategies {
		if s == v {
			return true
		}
	}
	return false
}

func promptIDStrategy() (string, error) {
	options := []huh.Option[string]{
		huh.NewOption("Sequential (001, 002, ...)", "sequential").Selected(true),
		huh.NewOption("Prefixed (dr-001, dr-002, ...)", "prefixed"),
		huh.NewOption("Random (a3f9x2, b7k2m1, ...)", "random"),
		huh.NewOption("ULID (01h5a3mpk, ...)", "ulid"),
	}

	var selected string
	err := huh.NewSelect[string]().
		Title("ID strategy").
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return "sequential", fmt.Errorf("prompt cancelled: %w", err)
	}
	if selected == "" {
		return "sequential", nil
	}
	return selected, nil
}

func resolveIDPrefix(cmd *cobra.Command, isTTY bool) (string, error) {
	if cmd.Flags().Changed("id-prefix") {
		if projectInitIDPrefix == "" {
			return "", fmt.Errorf("--id-prefix cannot be empty for prefixed strategy")
		}
		return projectInitIDPrefix, nil
	}
	if isTTY {
		value := ""
		err := huh.NewInput().
			Title("ID prefix").
			Description("Short prefix for task IDs (e.g., \"dr\", \"cli\")").
			Value(&value).
			Run()
		if err != nil {
			return "", fmt.Errorf("prompt cancelled: %w", err)
		}
		if value == "" {
			return "", fmt.Errorf("prefix cannot be empty for prefixed strategy")
		}
		return value, nil
	}
	return "", fmt.Errorf("--id-prefix is required for prefixed strategy in non-interactive mode")
}

// idStrategyExamples holds strategy-specific placeholder replacements.
type idStrategyExamples struct {
	exampleID       string // e.g. "001"
	exampleFilename string // e.g. "015-add-user-auth.md"
	filePattern     string // e.g. "NNN-descriptive-title.md"
}

func getIDStrategyExamples(cfg idStrategyConfig) idStrategyExamples {
	switch cfg.strategy {
	case "prefixed":
		p := cfg.prefix
		return idStrategyExamples{
			exampleID:       fmt.Sprintf("%s-001", p),
			exampleFilename: fmt.Sprintf("%s-015-add-user-auth.md", p),
			filePattern:     fmt.Sprintf("%s-NNN-descriptive-title.md", strings.ToUpper(p)),
		}
	case "random":
		return idStrategyExamples{
			exampleID:       "a3f9x2",
			exampleFilename: "a3f9x2-add-user-auth.md",
			filePattern:     "ID-descriptive-title.md",
		}
	case "ulid":
		return idStrategyExamples{
			exampleID:       "01h5a3mpk",
			exampleFilename: "01h5a3mpk-add-user-auth.md",
			filePattern:     "ID-descriptive-title.md",
		}
	default: // sequential
		return idStrategyExamples{
			exampleID:       "001",
			exampleFilename: "015-add-user-auth.md",
			filePattern:     "NNN-descriptive-title.md",
		}
	}
}

// applyIDStrategyReplacements replaces sequential-style placeholders in template content.
func applyIDStrategyReplacements(content []byte, examples idStrategyExamples) []byte {
	// Only apply if non-sequential (sequential is already the default in templates)
	if examples.exampleID == "001" {
		return content
	}
	result := content
	result = bytes.ReplaceAll(result, []byte(`id: "001"`), []byte(fmt.Sprintf(`id: "%s"`, examples.exampleID)))
	result = bytes.ReplaceAll(result, []byte("015-add-user-auth.md"), []byte(examples.exampleFilename))
	result = bytes.ReplaceAll(result, []byte("NNN-descriptive-title.md"), []byte(examples.filePattern))
	return result
}

// idStrategySpecSection returns a markdown section documenting the chosen ID strategy.
func idStrategySpecSection(cfg idStrategyConfig) string {
	switch cfg.strategy {
	case "prefixed":
		return fmt.Sprintf(`

## ID Generation

This project uses **prefixed sequential** IDs with the prefix **%q**.

- Format: `+"`%s-NNN`"+`
- IDs are zero-padded sequential numbers with a prefix (e.g., `+"`%s-001`"+`, `+"`%s-002`"+`)
- The prefix groups tasks by team, area, or project
`, cfg.prefix, cfg.prefix, cfg.prefix, cfg.prefix)
	case "random":
		return `

## ID Generation

This project uses **random** IDs.

- Format: alphanumeric strings (e.g., ` + "`a3f9x2`" + `, ` + "`b7k2m1`" + `)
- Generated automatically by ` + "`taskmd add`" + `
- Default length: 6 characters
`
	case "ulid":
		return `

## ID Generation

This project uses **ULID** (Universally Unique Lexicographically Sortable Identifier) IDs.

- Format: lowercase ULID strings (e.g., ` + "`01h5a3mpk`" + `)
- Generated automatically by ` + "`taskmd add`" + `
- ULIDs are time-ordered, so tasks sort chronologically by creation
- Default length: 9 characters
`
	default: // sequential
		return "" // sequential is the default, no extra section needed
	}
}

// collectInitFiles returns files split into root and task dir.
// Agent configs and spec are both placed in the task directory.
func collectInitFiles(idStrategy idStrategyConfig) (rootFiles, taskDirFiles []fileToWrite) {
	examples := getIDStrategyExamples(idStrategy)

	if !projectInitNoAgent {
		agents := getProjectInitAgents()
		for _, agent := range agents {
			taskDirFiles = append(taskDirFiles, fileToWrite{
				filename: agent.filename,
				content:  applyIDStrategyReplacements(agent.template, examples),
			})
		}
	}

	if !projectInitNoSpec {
		specContent := applyIDStrategyReplacements(initSpecTemplate, examples)
		// Append strategy-specific documentation section
		if section := idStrategySpecSection(idStrategy); section != "" {
			specContent = append(specContent, []byte(section)...)
		}
		taskDirFiles = append(taskDirFiles, fileToWrite{
			filename: specFilename,
			content:  specContent,
		})
	}

	return rootFiles, taskDirFiles
}

func getProjectInitAgents() []agentConfig {
	var agents []agentConfig

	// If no agent flags specified, default to Claude
	if !projectInitClaude && !projectInitGemini && !projectInitCodex {
		projectInitClaude = true
	}

	if projectInitClaude {
		agents = append(agents, agentConfig{
			name:     "Claude Code",
			filename: "CLAUDE.md",
			template: claudeTemplate,
		})
	}

	if projectInitGemini {
		agents = append(agents, agentConfig{
			name:     "Gemini",
			filename: "GEMINI.md",
			template: geminiTemplate,
		})
	}

	if projectInitCodex {
		agents = append(agents, agentConfig{
			name:     "Codex",
			filename: "AGENTS.md",
			template: codexTemplate,
		})
	}

	return agents
}

// ensureTaskDir creates the task directory if it doesn't exist.
func ensureTaskDir(path string, quiet bool) (created bool, err error) {
	info, statErr := os.Stat(path)
	if statErr == nil {
		if !info.IsDir() {
			return false, fmt.Errorf("not a directory: %s", path)
		}
		if !quiet {
			fmt.Fprintf(os.Stderr, "Task directory already exists: %s\n", path)
		}
		return false, nil
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return false, fmt.Errorf("failed to create task directory: %w", err)
	}
	return true, nil
}

// writeConfigFile writes .taskmd.yaml to the project root.
func writeConfigFile(root, taskDirPath string, idStrategy idStrategyConfig, quiet bool) (created bool, err error) {
	configPath := filepath.Join(root, configFilename)

	if !projectInitForce {
		if _, err := os.Stat(configPath); err == nil {
			if !quiet {
				abs, _ := filepath.Abs(configPath)
				fmt.Fprintf(os.Stderr, "Skipped %s (already exists, use --force to overwrite)\n", abs)
			}
			return false, nil
		}
	}

	content := fmt.Sprintf("dir: %s\n", taskDirPath)
	content += buildIDConfigYAML(idStrategy)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return false, fmt.Errorf("failed to write %s: %w", configFilename, err)
	}
	return true, nil
}

// buildIDConfigYAML returns the id: section for .taskmd.yaml.
// Returns empty string for the default sequential strategy.
func buildIDConfigYAML(cfg idStrategyConfig) string {
	switch cfg.strategy {
	case "prefixed":
		return fmt.Sprintf("id:\n  strategy: prefixed\n  prefix: %s\n", cfg.prefix)
	case "random":
		return "id:\n  strategy: random\n  length: 6\n"
	case "ulid":
		return "id:\n  strategy: ulid\n  length: 9\n"
	default:
		return ""
	}
}

// writeInitFiles writes files to a directory, returning created paths.
func writeInitFiles(dir string, files []fileToWrite, quiet bool) ([]string, error) {
	var created []string

	for _, f := range files {
		absPath, skipped, err := writeInitFile(dir, f)
		if err != nil {
			return created, err
		}
		if skipped {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Skipped %s (already exists, use --force to overwrite)\n", absPath)
			}
			continue
		}
		created = append(created, absPath)
	}

	return created, nil
}

func writeInitFile(targetDir string, f fileToWrite) (absPath string, skipped bool, err error) {
	outputPath := filepath.Join(targetDir, f.filename)
	absPath, err = filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	if !projectInitForce {
		if _, err := os.Stat(outputPath); err == nil {
			return absPath, true, nil
		}
	}

	if err := os.WriteFile(outputPath, f.content, 0644); err != nil {
		return absPath, false, fmt.Errorf("failed to write %s: %w", f.filename, err)
	}

	return absPath, false, nil
}

// printInitSummary prints the list of created files and next steps.
func printInitSummary(createdPaths []string) {
	if len(createdPaths) == 0 {
		fmt.Fprintln(os.Stderr, "Nothing to create (everything already exists).")
		return
	}

	fmt.Println("\nCreated:")
	for _, p := range createdPaths {
		fmt.Printf("  %s\n", p)
	}
	fmt.Println("\nYou're ready! Try:")
	fmt.Println("  taskmd add \"My first task\"")
	fmt.Println("  taskmd list")
	fmt.Println("  taskmd web start --open")
}

func printFilesToStdout(files []fileToWrite) error {
	for i, f := range files {
		if i > 0 {
			fmt.Print("\n---\n")
			fmt.Printf("# %s\n", f.filename)
			fmt.Print("---\n\n")
		}
		fmt.Print(string(f.content))
	}
	return nil
}
