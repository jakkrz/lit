package cmd

import (
	"fmt"
	"lit/index"

	"github.com/spf13/cobra"
)

var (
	Add = cobra.Command{
		Use:   "add <file-or-folder>",
		Short: "adds things to the index",
		Long:  "recursively adds files of the folder to the index",
		Run: func(cmd *cobra.Command, args []string) {
			if !IsRepo() {
				fmt.Println("fatal: not a repository!")
				return
			}

			path := args[0]

			err := index.Stage(path)

			if err != nil {
				fmt.Println(err)
				return
			}
		},
		Args: cobra.MaximumNArgs(1),
	}
)

func init() {
	RootCmd.AddCommand(&Add)
}
