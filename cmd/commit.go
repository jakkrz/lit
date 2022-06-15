package cmd

import (
	"fmt"
	"lit/index"

	"github.com/spf13/cobra"
)

var (
	Commit = cobra.Command{
		Use:   "commit",
		Short: "shows commits",
		Long:  "shows every commit upstream of the current commit",
		Run: func(_ *cobra.Command, args []string) {
			if !IsRepo() {
				fmt.Println("fatal: not a repository!")
				return
			}
			
			_, err := index.Commit(args[0])

			if err != nil {
				fmt.Println("Couldn't commit:", err)
			}
		},
		Args: cobra.ExactArgs(1),
	}
)

func init() {
	RootCmd.AddCommand(&Commit)
}
