// package refs implements functions for manipulating lit
// refrences, i.e. branches and HEAD.
package refs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"lit/objects"
	"lit/util"
	"os"
)

// BranchExists checks if a branch exists with the given name, returning
// an error if os.Stat fails.
func BranchExists(name string) (bool, error) {
	_, err := os.Stat(".lit/refs/heads/" + name)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, errors.New("failed to check if branch exists")
}

// CreateBranchTo creates a branch with a given name to a given commit hash.
func CreateBranchTo(name string, hash string) error {
	exists, err := BranchExists(name)

	if err != nil {
		return err
	}

	if exists {
		return errors.New("branch already exists")
	}

	isCommit := objects.HashIsCommit(hash)

	if !isCommit {
		return fmt.Errorf("%s is not the hash of a commit", hash)
	}

	err = util.WriteJSON(".lit/refs/heads/"+name, BranchContent{hash})

	if err != nil {
		return err
	}

	return nil
}

func SetBranchTo(name string, hash string) error {
	return util.WriteJSON(".lit/refs/heads/"+name, BranchContent{hash})
}

// DeleteBranch deletes a branch, returning an error if the branch doesn't exist.
func DeleteBranch(name string) error {
	exists, err := BranchExists(name)

	if err != nil {
		return err
	}

	if !exists {
		return ErrNotFound
	}

	err = os.Remove(".lit/refs/heads/" + name)

	if err != nil {
		return err
	}

	return nil
}

func DeleteBranchSafe(name string) error {
	hc, err := ReadHead()

	if err != nil {
		return err
	}

	if !hc.Detached && hc.Location == name {
		loc, err := HeadCommit()

		if err != nil {
			return err
		}

		SetHeadTo(HeadContent{Detached: true, Location: loc})
	}

	err = DeleteBranch(name)

	if err != nil {
		return err
	}

	return nil
}

type BranchContent struct {
	Reference string
}

func ReadBranch(name string) (string, error) {
	exists, err := BranchExists(name)

	if err != nil {
		return "", err
	}

	if !exists {
		return "", ErrNotFound
	}

	data, err := os.ReadFile(".lit/refs/heads/" + name)

	if err != nil {
		if err == os.ErrNotExist {
			return "", ErrNotFound
		}
		
		return "", err
	}

	var content BranchContent

	err = json.Unmarshal(data, &content)

	if err != nil {
		return "", err
	}

	return content.Reference, nil
}

func GetBranchNames() []string {
	names := make([]string, 0, 1)

	err := util.ForeachSubfile(".lit/refs/heads", func(path string, d fs.DirEntry) error {
		names = append(names, d.Name())

		return nil
	})

	if err != nil {
		return nil
	}
	
	return names
}