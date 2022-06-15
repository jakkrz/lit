/*
Package index implements I/O functions to manipulate the index file.
Functions manipulating the index will often work with map[string]strings,
as the index contains pairs of file paths and hashes of blobified versions
of the files. These pairs are simply refered to as 'pairs'.
The terms 'staging area' and 'to stage' are synonymous to 'index'
and 'add to index', respectively.
*/
package index

import (
	"errors"
	"io/fs"
	"lit/objects"
	"lit/refs"
	"lit/set"

	"lit/util"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"
	"time"
)

// SetDefault initializes the index to its default state containing no
// paths pointing to blob hashes.
func SetDefault() error {
	return util.WriteJSON(".lit/index", map[string]string{})
}

// StagePairs adds pairs to the index.
func StagePairs(pairs map[string]string) error {
	staged, err := Staged()

	if err != nil {
		return err
	}

	for path, blobHash := range pairs {
		staged[path] = blobHash
	}

	return SetStaged(staged)
}

// SetStaged overrides the content of the index to the specified pairs.
func SetStaged(pairs map[string]string) error {
	err := util.WriteJSON(".lit/index", pairs)

	if err != nil {
		return err
	}

	return nil
}

// Staged returns the contents of the index
func Staged() (map[string]string, error) {
	result := map[string]string{}
	err := util.ReadJSON(".lit/index", &result)

	return result, err
}

// satisfyUnstagedChanges modifies the staged map so that no more unstaged
// changes of filepaths that pass the predicate are reported to
// be changes in the future.
func satisfyUnstagedChanges(predicate func(string) bool, unstagedChanges map[string]Status, staged map[string]string) error {
	for filepathChanged, state := range unstagedChanges {
		if !predicate(filepathChanged) {
			continue
		}

		switch state {
		case Modified:
			hash := objects.Blobify(filepathChanged)

			if hash == "" {
				return errors.New("couldn't write files to objects")
			}

			staged[filepathChanged] = hash

		case Deleted:
			delete(staged, filepathChanged)
		default:
			panic(state) // An unexpected value for the state of an unstaged change
		}
	}

	return nil
}

// satisfyUntracked blobifies untracked files that pass the predicate.
func satisfyUntracked(predicate func(string) bool, untracked []string, staged map[string]string) error {
	for _, untrackedPath := range untracked {
		if predicate(untrackedPath) {
			hash := objects.Blobify(untrackedPath)

			if hash == "" {
				return errors.New("couldn't write files to objects")
			}

			staged[untrackedPath] = hash
		}
	}

	return nil
}

// Stage blobifies and adds the given file at the path to the index.
// If the path points to a directory, Stage recursively applies
// the process to sub-files.
func Stage(path string) error {
	staged, err := Staged()

	if err != nil {
		return err
	}

	// we keep track if any changes were made to the index, reporting an error if none were made
	changesDone := false

	unstagedChanges, untracked, err := UnstagedChanges(staged)

	if err != nil {
		return err
	}

	// if any changes are being done the predicate will set changesDone to true
	predicate := func(s string) bool {
		if util.IsSubPath(path, s) {
			changesDone = true
			return true
		} else {
			return false
		}
	}

	if err = satisfyUnstagedChanges(predicate, unstagedChanges, staged); err != nil {
		return err
	}

	if err = satisfyUntracked(predicate, untracked, staged); err != nil {
		return err
	}

	if !changesDone {
		return errors.New("file does not exist")
	}

	if err = SetStaged(staged); err != nil {
		return err
	}

	return nil
}

type Status uint8

const (
	Created Status = iota
	Modified
	Deleted
	Renamed
)

func (s Status) String() string {
	switch s {
	case Created:
		return "created"
	case Modified:
		return "modified"
	case Deleted:
		return "deleted"
	case Renamed:
		return "renamed"
	default:
		return "???"
	}
}

type treeNode struct {
	subtrees map[string]treeNode
	blobs    map[string]string
}

func newTreeNode() treeNode {
	return treeNode{subtrees: map[string]treeNode{}, blobs: map[string]string{}}
}

// generateTreeNode generates a treeNode, taking into account subdirectories specified by slashes
// and from that creating a tree structure.
func generateTreeNode(hashes map[string]string) treeNode {
	rootTree := newTreeNode()

	for path, hash := range hashes {
		dir, base := pathlib.Split(path)
		dir = strings.TrimSuffix(dir, "/")

		currentTree := rootTree

		if dir != "" {
			for _, subdir := range strings.Split(dir, "/") {
				newTreeN := newTreeNode()
				currentTree.subtrees[subdir] = newTreeN
				currentTree = newTreeN
			}
		}

		currentTree.blobs[base] = hash
	}

	return rootTree
}

// Write writes the treeNode into lit/objects, returning the hash of the upmost-level tree.
func (tn treeNode) Write() string {
	hashes := map[string]objects.TreeEntry{}

	for name, sub := range tn.subtrees {
		hashes[name] = objects.TreeEntry{ObjType: "Tree", Hash: sub.Write()}
	}

	for name, blobHash := range tn.blobs {
		hashes[name] = objects.TreeEntry{ObjType: "Blob", Hash: blobHash}
	}

	return objects.WriteTree(hashes)
}

// Commit creates a commit with the given name.
func Commit(commitName string) (string, error) {
	hashes, err := Staged()

	if err != nil {
		return "", err
	}

	tree := generateTreeNode(hashes).Write()

	if tree == "" {
		return "", errors.New("failed to write commit")
	}

	prevHead, err := refs.HeadCommit()

	if err != nil {
		if errors.Is(err, refs.ErrNotFound) {
			prevHead = ""
		} else {
			return "", err
		}
	}

	commitStruct := objects.NewCommit(commitName, tree, time.Now())
	commitStruct.Parents = []string{prevHead}
	com := objects.WriteCommit(commitStruct)

	if com == "" {
		return "", errors.New("failed to write commit")
	}

	err = refs.NudgeHead(com)

	if err != nil {
		return "", err
	}

	return com, nil
}

// UnstagedChanges returns a map of unstaged changes and a slice of untracked files.
func UnstagedChanges(staged map[string]string) (map[string]Status, []string, error) {
	staged = copyMap(staged)

	untracked := []string{}
	unstagedResult := map[string]Status{}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		cleanPath := util.CleanPath(path)

		if cleanPath == ".lit" {
			return filepath.SkipDir
		}

		if cleanPath == "." || d.IsDir() {
			return nil
		}

		data, err := os.ReadFile(cleanPath)

		if err != nil {
			return err
		}

		hash := objects.Hash(data)

		indexHash, exists := staged[cleanPath]

		if !exists {
			untracked = append(untracked, cleanPath)
			delete(staged, cleanPath)

			return nil
		}

		if hash != indexHash {
			unstagedResult[cleanPath] = Modified
		}

		delete(staged, cleanPath)

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	for path := range staged {
		unstagedResult[util.CleanPath(path)] = Deleted
	}

	return unstagedResult, untracked, nil
}

// recursivelyGetHashes recursively gets hashes and inserts them into the
// result map with the given basePath.
func recursivelyGetHashes(hash string, result map[string]string, basePath string) error {
	tree, err := objects.ReadAsTree(hash)

	if err != nil {
		return err
	}

	for name, entry := range tree {
		if entry.ObjType == "Tree" {
			if err = recursivelyGetHashes(entry.Hash, result, basePath+name+"/"); err != nil {
				return err
			}
		} else {
			result[basePath+name] = entry.Hash
		}
	}

	return nil
}

// StagedChanges returns a map of changes to files compared to the previous commit.
func StagedChanges(hashes map[string]string) (map[string]Status, error) {
	hashes = copyMap(hashes)

	stagedStatus := map[string]Status{}

	headCommit, err := refs.HeadCommit()

	if err != nil {
		if errors.Is(err, refs.ErrNotFound) {
			for path := range hashes {
				stagedStatus[path] = Created
			}

			return stagedStatus, nil
		}

		return nil, err
	}

	com, err := objects.ReadAsCommit(headCommit)

	if err != nil {
		return nil, err
	}

	commitHashes := map[string]string{}
	recursivelyGetHashes(com.CommitTree, commitHashes, "")

	for name, commitHash := range commitHashes {
		indexHash, exists := hashes[name]

		if !exists {
			stagedStatus[name] = Deleted
			delete(hashes, name)
			continue
		}

		if commitHash != indexHash {
			stagedStatus[name] = Modified
		}

		delete(hashes, name)
	}

	for name := range hashes {
		stagedStatus[name] = Created
	}

	return stagedStatus, nil
}

func copyMap(original map[string]string) map[string]string {
	targetMap := make(map[string]string)

	// Copy from the original map to the target map
	for key, value := range original {
		targetMap[key] = value
	}

	return targetMap
}

// GetStatus compares the contents of the index, working-tree and previous commit
// to produce a list of changes to files
func GetStatus() (map[string]Status, map[string]Status, []string, error) {

	hashes, err := Staged()

	if err != nil {
		return nil, nil, nil, err
	}

	unstagedStatus, untracked, err := UnstagedChanges(hashes)

	if err != nil {
		return nil, nil, nil, err
	}

	stagedStatus, err := StagedChanges(hashes)

	if err != nil {
		return nil, nil, nil, err
	}

	return unstagedStatus, stagedStatus, untracked, nil
}

// ClearWorkingTree clears the working tree leaving certain files untouched.
func ClearWorkingTree(leaveAlone set.Set[string]) error {
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		cleanPath := util.CleanPath(path)

		if cleanPath == "." {
			return nil
		}

		if cleanPath == ".lit" {
			return filepath.SkipDir
		}

		if _, exists := leaveAlone[path]; !exists {
			err = os.RemoveAll(path)

			if err != nil {
				return err
			}

			if d.IsDir() {
				return filepath.SkipDir
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// LoadIn loads in a commit.
func LoadIn(com *objects.Commit) error {
	tree, err := objects.ReadAsTree(com.CommitTree)

	if err != nil {
		return err
	}

	err = objects.LoadTree("", tree)

	if err != nil {
		return err
	}

	return nil
}

// recursivelyAddToIndex recursively adds elements of the tree into the index map.
func recursivelyAddToIndex(tree map[string]objects.TreeEntry, index map[string]string, basePath string) error {
	for name, entry := range tree {
		if entry.ObjType == "Tree" {
			subTree, err := objects.ReadAsTree(entry.Hash)

			if err != nil {
				return err
			}

			err = recursivelyAddToIndex(subTree, index, basePath+name+"/")

			if err != nil {
				return err
			}

			continue
		}

		index[basePath+name] = entry.Hash
	}

	return nil
}

func LoadIntoIndex(stagedChanges map[string]Status, hashes map[string]string, commit *objects.Commit) error {
	tree, err := objects.ReadAsTree(commit.CommitTree)

	if err != nil {
		return err
	}

	result := map[string]string{}

	for path := range stagedChanges {
		result[path] = hashes[path]
	}

	err = recursivelyAddToIndex(tree, result, "")

	if err != nil {
		return err
	}

	return SetStaged(result)
}
