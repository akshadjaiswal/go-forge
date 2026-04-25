package generator

import (
	"fmt"
	"os"
	"os/exec"
)

// PostGenerate runs post-generation steps inside the created project directory:
//  1. go mod tidy   — resolves and pins all dependencies
//  2. git init      — initializes a git repo
//  3. git add + commit — initial commit so the project starts with clean history
func PostGenerate(opts ProjectOptions) error {
	dir := opts.Name

	fmt.Println()

	if err := runInDir(dir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	fmt.Println("✓ go mod tidy")

	// git init -b main requires Git ≥ 2.28. Fall back gracefully on older versions.
	if err := runInDir(dir, "git", "init", "-b", "main"); err != nil {
		if err2 := runInDir(dir, "git", "init"); err2 != nil {
			return fmt.Errorf("git init: %w", err2)
		}
		// Best-effort rename — ignore error if branch already named main or rename unsupported.
		_ = runInDir(dir, "git", "branch", "-m", "main")
	}
	fmt.Println("✓ git init")

	if err := runInDir(dir, "git", "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	if err := runInDir(dir, "git", "commit", "-m", "chore: initial scaffold from forge"); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	fmt.Println("✓ initial commit")

	fmt.Println()
	fmt.Printf("✓ Created %s/\n", opts.Name)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", opts.Name)
	fmt.Printf("  cp .env.example .env    # fill in DB_URL and JWT_SECRET\n")
	fmt.Printf("  make dev                # start the server\n")

	return nil
}

// runInDir runs a command with its working directory set to dir.
func runInDir(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
