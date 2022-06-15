// package refs implements functions for manipulating lit refrences, i.e. branches and HEAD.
package refs

import (
	"errors"
	"lit/objects"
	"lit/util"
	"strings"
)

// HeadContent represents the JSON content of HEAD in struct form.
type HeadContent struct {
	// Detached specifies whether HEAD points to a commit (true) or a branch (false).
	Detached bool
	// Location specifies the commit hash or branch name that HEAD points to.
	Location string
}

var (
	ErrCouldNotRead = errors.New("could not read")
)

// ReadHead reads the content of HEAD and returns it in a struct.
func ReadHead() (HeadContent, error) {
	var hc HeadContent
	err := util.ReadJSON(".lit/HEAD", &hc)

	if err != nil {
		return HeadContent{}, ErrCouldNotRead
	}

	return hc, nil
}

// HeadCommit returns hash of commit pointed to by HEAD or by the current branch.
func HeadCommit() (string, error) {
	hc, err := ReadHead()

	if err != nil {
		return "", err
	}

	if hc.Location == "" {
		return "", nil
	}

	if hc.Detached {
		return hc.Location, nil
	}

	hash, err := ReadBranch(hc.Location)

	if err != nil {
		return "", err
	}

	return hash, nil
}

func NudgeHead(commitHash string) error {
	headContent, err := ReadHead()

	if err != nil {
		return err
	}

	if headContent.Location == "" {
		return nil
	}

	if headContent.Detached {
		headContent.Location = commitHash
		SetHeadTo(headContent)
	} else {
		exists, err := BranchExists(headContent.Location)

		if err != nil {
			return err
		}

		if !exists {
			err = CreateBranchTo(headContent.Location, commitHash)

			if err != nil {
				return err
			}
		}

		err = SetBranchTo(headContent.Location, commitHash)

		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// SetHeadTo sets HEAD to the specified content, checking for invalid locations and returning ErrCouldNotFind on failure.
func SetHeadTo(hc HeadContent) error {
	// verify the location
	if hc.Detached {
		isCommit := objects.HashIsCommit(hc.Location)
		if !isCommit {
			return ErrNotFound
		}
	} else {
		isBranch, err := BranchExists(hc.Location)

		if err != nil || !isBranch {
			return ErrNotFound
		}
	}

	return util.WriteJSON(".lit/HEAD", hc)
}

// ErrNotFound is a sentinel error for locations not found.
var ErrNotFound = errors.New("could not find location")

// SetHeadToString sets HEAD to a branch if location is a branch name, otherwise it sets it to a commit hash.
func SetHeadToString(location string) error {
	exists, err := BranchExists(location)

	if err != nil {
		return err
	}

	if exists {
		SetHeadTo(HeadContent{Detached: false, Location: location})
		return nil
	}

	lowerLocation := strings.ToLower(location)

	// search commits for hash
	hash, err := objects.ExpandHash(lowerLocation)

	if err != nil {
		return err
	}

	if hash == "" {
		return ErrNotFound
	}

	err = SetHeadTo(HeadContent{Detached: true, Location: hash})

	if err != nil {
		return err
	}

	return nil
}
