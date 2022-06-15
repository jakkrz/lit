package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"lit/index"
	"lit/refs"
	"lit/util"
	"os"

	"github.com/spf13/cobra"
)

// ErrInitFileCreation is returned by InitializeRepo when it cannot create a file.
var ErrInitFileCreation = errors.New("failed to instantiate required files")

// InitializeRepo initializes a new repo if possible, otherwise returns an error.
func InitializeRepo() error {
	err := os.Mkdir(".lit", 0777)

	if errors.Is(err, fs.ErrExist) {
		return errors.New("already a lit repository")
	} else if err != nil {
		return ErrInitFileCreation
	}

	err = os.Mkdir(".lit/objects", 0777)

	if err != nil {
		return ErrInitFileCreation
	}

	err = os.MkdirAll(".lit/refs/heads", 0777)

	if err != nil {
		return ErrInitFileCreation
	}

	err = util.WriteJSON(".lit/HEAD", refs.HeadContent{Detached: false, Location: "main"})

	if err != nil {
		return ErrInitFileCreation
	}

	err = index.SetDefault()

	if err != nil {
		return ErrInitFileCreation
	}

	return nil
}

var (
	Init = cobra.Command{
		Use:   "init",
		Short: "initialize a repository",
		Long:  "initializes a repository in the .lit file of the current working directory",
		Run: func(_ *cobra.Command, _ []string) {
			if err := InitializeRepo(); err != nil {
				fmt.Println("Could not initialize:", err)
			} else {
				fmt.Println("Successfully initialized empty lit repository. Happy coding!")
			}
		},
		Args: cobra.NoArgs,
	}
)

func init() {
	RootCmd.AddCommand(&Init)
}
