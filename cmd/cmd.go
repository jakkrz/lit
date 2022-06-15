package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	RootCmd = cobra.Command{
		Use:   "lit",
		Short: "a git clone",
		Long:  "a minimalistic version management system that works similarly to git",
	}
)

// IsRepo checks if the current working directory contains a .lit folder.
func IsRepo() bool {
	stat, err := os.Stat("./.lit/")
	return err == nil && stat.IsDir()
}

// Execute executes the root command
func Execute() error {
	return RootCmd.Execute()
}
