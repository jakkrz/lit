package cmd

import (
	"errors"
	"fmt"
	"lit/refs"

	"github.com/spf13/cobra"
)

func DisplayBranches() {
	headContent, err := refs.ReadHead()

	if err != nil {
		fmt.Println(err)
		return
	}

	names := refs.GetBranchNames()

	if names == nil {
		fmt.Println("Error while reading branches")
	}

	for _, branch := range names {
		fmt.Print(branch)

		if !headContent.Detached && branch == headContent.Location {
			fmt.Print(" *")
		}
		fmt.Print("\n")
	}
}

var (
	Branch = cobra.Command{
		Use:   "branch <name>",
		Short: "manipulates branches",
		Long:  "creates a new branch with name <name> if provided, else lists branches",
		Run: func(cmd *cobra.Command, args []string) {
			if !IsRepo() {
				fmt.Println("fatal: not a repository!")
				return
			}

			if len(args) == 0 {
				DisplayBranches()
				return
			}

			
			deleteFlag, err := cmd.Flags().GetBool("delete")

			if err != nil {
				panic(err)
			}
			
			name := args[0]
			
			if deleteFlag {
				err = refs.DeleteBranchSafe(name)

				if err != nil {
					fmt.Println(err)
				}
			} else {
				headCommit, err := refs.HeadCommit()

				if err != nil {
					if errors.Is(err, refs.ErrNotFound) {
						fmt.Println("Could not create branch: head does not point to a commit yet")
						return
					}

					fmt.Println(err)
					return
				}

				err = refs.CreateBranchTo(name, headCommit)

				if err != nil {
					fmt.Println(err)
				}
			}
		},
		Args: cobra.MaximumNArgs(1),
	}
)

func init() {
	RootCmd.AddCommand(&Branch)
	Branch.Flags().BoolP("delete", "d", false, "deletes branch instead of creating it")
}
