package cmd

import (
	"errors"
	"fmt"
	"lit/objects"
	"lit/refs"

	"lit/index"
	"lit/set"

	"github.com/spf13/cobra"
)

func switchTo(hc refs.HeadContent) error {
	currentHashes, err := index.Staged()

	if err != nil {
		return err
	}

	currentUnstagedStatus, currentStagedStatus, currentUntracked, err := index.GetStatus()

	if err != nil {
		return err
	}

	if len(currentUnstagedStatus)+len(currentStagedStatus) != 0 {
		return errors.New("uncommitted changes")
	}

	if err = index.ClearWorkingTree(set.FromSlice(currentUntracked)); err != nil {
		return err
	}

	if err = refs.SetHeadTo(hc); err != nil {
		return err
	}

	newCommitHash, err := refs.HeadCommit()

	if err != nil {
		return err
	}

	newCommitContent, err := objects.ReadAsCommit(newCommitHash)

	if err != nil {
		return err
	}

	if err = index.LoadIntoIndex(currentStagedStatus, currentHashes, newCommitContent); err != nil {
		return err
	}

	if err = index.LoadIn(newCommitContent); err != nil {
		return err
	}

	return nil
}

func checkoutLocation(loc string, detach bool) error {
	commit, err := refs.ReadBranch(loc)

	if detach {
		if err != nil {
			return err
		}

		if err = switchTo(refs.HeadContent{Detached: true, Location: commit}); err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		if errors.Is(err, refs.ErrNotFound) {
			hash, err := objects.ExpandHash(loc)

			if err != nil {
				return err
			}

			if hash == "" {
				return errors.New("couldn't find commit with hash")
			}

			if err = switchTo(refs.HeadContent{Detached: true, Location: hash}); err != nil {
				return err
			}

			return nil
		} else {
			return err
		}
	}

	switchTo(refs.HeadContent{Detached: false, Location: loc})

	return nil
}

var (
	Checkout = cobra.Command{
		Use:   "checkout <location>",
		Short: "points HEAD to location",
		Long:  "points HEAD to branch. If location is not a branch, checkout searches objects",
		Run: func(cmd *cobra.Command, args []string) {
			if !IsRepo() {
				fmt.Println("fatal: not a repository!")
				return
			}

			detach, err := cmd.Flags().GetBool("detach")

			if err != nil {
				panic(err)
			}

			loc := args[0]

			if err = checkoutLocation(loc, detach); err != nil {
				fmt.Println(err)
			}
		},
		Args: cobra.ExactArgs(1),
	}
)

func init() {
	RootCmd.AddCommand(&Checkout)
	Checkout.Flags().BoolP("detach", "d", false, "specify whether to checkout to the commit pointed by branch")
}
