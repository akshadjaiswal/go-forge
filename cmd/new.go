package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/akshadjaiswal/go-forge/internal/generator"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// flags for non-interactive mode
var (
	flagModule    string
	flagNoAuth    bool
	flagNoDocker  bool
)

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Scaffold a new production-ready Go API project",
	Long: `Creates a new Go REST API project with:
  • chi router + structured middleware
  • PostgreSQL + sqlx (no ORM)
  • JWT authentication (optional, --no-auth to skip)
  • slog structured logging
  • go-playground/validator
  • Dockerfile multi-stage build + docker-compose (optional, --no-docker to skip)
  • Makefile, .env.example, requests.http

Examples:
  forge new my-api
  forge new my-api --module github.com/user/my-api --no-auth`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

func init() {
	newCmd.Flags().StringVar(&flagModule, "module", "", "Go module path (e.g. github.com/user/my-api)")
	newCmd.Flags().BoolVar(&flagNoAuth, "no-auth", false, "Skip JWT authentication files")
	newCmd.Flags().BoolVar(&flagNoDocker, "no-docker", false, "Skip Docker/docker-compose files")
}

func runNew(cmd *cobra.Command, args []string) error {
	opts := generator.ProjectOptions{
		GoVersion:  generator.DetectGoVersion(),
		WithAuth:   !flagNoAuth,
		WithDocker: !flagNoDocker,
	}

	// 1. Project name — from arg or prompt
	if len(args) > 0 {
		opts.Name = strings.TrimSpace(args[0])
		if err := validateProjectName(opts.Name); err != nil {
			return err
		}
	} else {
		name, err := prompt("Project name", "", validateProjectName)
		if err != nil {
			return err
		}
		opts.Name = name
	}

	// Guard: don't overwrite existing directory
	if _, err := os.Stat(opts.Name); err == nil {
		return fmt.Errorf("directory %q already exists", opts.Name)
	}

	// 2. Module path — from flag or prompt
	if flagModule != "" {
		if err := validateModulePath(flagModule); err != nil {
			return err
		}
		opts.ModulePath = flagModule
	} else {
		defaultModule := fmt.Sprintf("github.com/username/%s", opts.Name)
		modulePath, err := prompt("Go module path", defaultModule, validateModulePath)
		if err != nil {
			return err
		}
		opts.ModulePath = modulePath
	}

	// 3. Feature toggles — prompt only when running interactively (no --module flag).
	// When --module is provided, assume non-interactive and use flag defaults.
	if flagModule == "" {
		withAuth, err := confirm("Include JWT authentication?", true)
		if err != nil {
			return err
		}
		opts.WithAuth = withAuth

		withDocker, err := confirm("Include Docker support?", true)
		if err != nil {
			return err
		}
		opts.WithDocker = withDocker
	}

	// Summary
	fmt.Println()
	fmt.Printf("  Project:  %s\n", opts.Name)
	fmt.Printf("  Module:   %s\n", opts.ModulePath)
	fmt.Printf("  Auth:     %v\n", opts.WithAuth)
	fmt.Printf("  Docker:   %v\n", opts.WithDocker)
	fmt.Printf("  Go:       %s\n", opts.GoVersion)
	fmt.Println()

	fmt.Printf("Generating %s...\n", opts.Name)

	// 4. Generate all files
	if err := generator.Generate(opts); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// 5. Post-generation: go mod tidy + git init + initial commit
	return generator.PostGenerate(opts)
}

// prompt shows an interactive text input with an optional default value.
func prompt(label, defaultVal string, validate promptui.ValidateFunc) (string, error) {
	p := promptui.Prompt{
		Label:    label,
		Default:  defaultVal,
		Validate: validate,
	}
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("prompt cancelled")
	}
	return strings.TrimSpace(result), nil
}

// confirm shows an interactive yes/no prompt.
func confirm(label string, defaultYes bool) (bool, error) {
	defaultStr := "Y/n"
	if !defaultYes {
		defaultStr = "y/N"
	}

	p := promptui.Prompt{
		Label:   fmt.Sprintf("%s (%s)", label, defaultStr),
		Default: map[bool]string{true: "Y", false: "N"}[defaultYes],
	}
	result, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("prompt cancelled")
	}
	result = strings.TrimSpace(strings.ToLower(result))
	if result == "" {
		return defaultYes, nil
	}
	return result == "y" || result == "yes", nil
}

var modulePathRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*/`)

// validateProjectName ensures the name is safe to use as a directory name.
func validateProjectName(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if s == "main" {
		return fmt.Errorf("project name cannot be 'main'")
	}
	if strings.Contains(s, "..") {
		return fmt.Errorf("project name cannot contain '..'")
	}
	if strings.ContainsAny(s, " \t/\\") {
		return fmt.Errorf("project name cannot contain spaces or slashes")
	}
	return nil
}

// validateModulePath ensures the module path is a valid Go module path.
func validateModulePath(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("module path cannot be empty")
	}
	if strings.Contains(s, "..") {
		return fmt.Errorf("module path cannot contain '..'")
	}
	if strings.ContainsAny(s, " \t\\") {
		return fmt.Errorf("module path cannot contain spaces")
	}
	if !modulePathRe.MatchString(s) {
		return fmt.Errorf("module path must start with a domain (e.g. github.com/user/repo)")
	}
	return nil
}
