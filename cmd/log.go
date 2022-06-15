package cmd

import (
	"errors"
	"fmt"
	"lit/objects"
	"lit/refs"
	"sort"

	"github.com/spf13/cobra"
)

type CommitHashPair struct {
	Com *objects.Commit
	Hash string
}

func recursivelyAddCommitsToSlice(commitHash string, out *[]CommitHashPair) error {
	commit, err := objects.ReadAsCommit(commitHash)

	if err != nil {
		return err
	}

	*out = append(*out, CommitHashPair{commit, commitHash})

	for _, parent := range commit.Parents {
		err = recursivelyAddCommitsToSlice(parent, out)

		if err != nil {
			return err
		}
	}

	return nil
}

type byTime []CommitHashPair

func (t byTime) Len() int {
	return len(t)
}

func (t byTime) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t byTime) Less(i, j int) bool {
	return t[i].Com.Time.After(t[j].Com.Time)
}

var (
	Log = cobra.Command{
		Use:   "log",
		Short: "shows commits",
		Long:  "shows every commit upstream of the current commit",
		Run: func(_ *cobra.Command, _ []string) {
			if !IsRepo() {
				fmt.Println("Not a repository!")
				return
			}

			commitHash, err := refs.HeadCommit()

			if err != nil {
				if errors.Is(err, refs.ErrNotFound) {
					fmt.Println("HEAD doesn't point to a commit yet.")
					return
				}

				fmt.Println(err)
				return
			}

			commits := []CommitHashPair{}

			err = recursivelyAddCommitsToSlice(commitHash, &commits)

			if errors.Is(err, objects.ErrNotOfType) {
				fmt.Printf("Error when back-tracking commits: not a commit!\n")
				return
			}
		
			if errors.Is(err, objects.ErrCouldNotRead) {
				fmt.Printf("Error when back-tracking commits: could not read a commit!\n")
				return
			}
		
			if err != nil {
				fmt.Printf("Unexpected error reading commit: %s\n", err)
				return
			}

			sort.Sort(byTime(commits))

			for _, commitHashPair := range commits {
				fmt.Println(commitHashPair.Hash, commitHashPair.Com.Name)
			}

		},
		Args: cobra.NoArgs,
	}
)

func init() {
	RootCmd.AddCommand(&Log)
}
