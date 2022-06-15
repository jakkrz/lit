package cmd

import (
	"fmt"
	"lit/index"

	"github.com/spf13/cobra"
)

var (
	Status = cobra.Command{
		Use:   "status",
		Short: "initialize a repository",
		Long:  "initializes a repository in the .lit file of the current working directory",
		Run: func(_ *cobra.Command, _ []string) {
			if !IsRepo() {
				fmt.Println("fatal: not a repository!")
				return
			}
			
			unstagedStatus, stagedStatus, untracked, err := index.GetStatus()

			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Untracked files:")

			for _, path := range untracked {
				fmt.Printf("\t%s\n", path)
			}

			fmt.Println("Changes not staged for commit:")
			
			for path, stat := range unstagedStatus {
				fmt.Printf("\t%s: %s\n", path, stat)
			}

			fmt.Println("Changes to be committed:")

			for path, stat := range stagedStatus {
				fmt.Printf("\t%s: %s\n", path, stat)
			}
		},
		Args: cobra.NoArgs,
	}
)

func init() {
	RootCmd.AddCommand(&Status)
}
