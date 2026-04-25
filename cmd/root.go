package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "forge — scaffold production-ready Go API projects",
	Long: `forge generates opinionated Go REST API projects with chi router,
sqlx + PostgreSQL, JWT auth, slog logging, Docker, and more.

Usage:
  forge new my-api`,
}

func Execute(version string) {
	rootCmd.Version = version
	rootCmd.AddCommand(newCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
