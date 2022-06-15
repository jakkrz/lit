package objects

import (
	"errors"
	"fmt"
	"io/fs"
	"lit/util"
	"os"
	"path/filepath"
	"strings"
)

// FolderCharacters specifies how many characters of the object's hash go into
// the folder name. The rest of the characters are the filename.
const FolderCharacters = 2

var (
	ErrNotOfType    = errors.New("not of the required type")
	ErrCouldNotRead = errors.New("could not read file")
)

func LoadTree(basePath string, tree map[string]TreeEntry) error {
	for name, entry := range tree {
		if entry.ObjType == "Tree" {
			subtree, err := ReadAsTree(entry.Hash)

			if err != nil {
				return err
			}

			if len(subtree) != 0 {
				err = os.MkdirAll(basePath+name, 0777)

				if err != nil {
					return err
				}
			}

			err = LoadTree(basePath+name+"/", subtree)

			if err != nil {
				return err
			}

			continue
		}

		blob, err := ReadAsBlob(entry.Hash)

		if err != nil {
			return err
		}

		err = os.WriteFile(basePath+name, []byte(blob), 0777)

		if err != nil {
			return err
		}

		fmt.Println("created", basePath+name)
	}

	return nil
}


// concatLastPathElems gets the n last slash-seperated elements of a path and concatenates them. E.g. if
// path is foo/bar/baz, and n is 2, the result is barbaz. If n is greater than
// the amount of elements in path, concatLastPathElems will return every element of the path
// concatenated together.
func concatLastPathElems(path string, n int) string {
	elements := strings.Split(path, "/")

	if n > len(elements) {
		return strings.Join(elements, "")
	} else {
		return strings.Join(elements[len(elements)-n:], "")
	}
}

func ExpandHash(hash string) (string, error) {
	hash = strings.ToLower(hash)
	result := ""

	err := filepath.Walk(".lit/objects/", func(path string, info fs.FileInfo, err error) error {
		cleanPath := util.CleanPath(path)
		if info.IsDir() {
			if cleanPath == ".lit/objects" {
				return nil
			}

			if info.Name() != hash[:FolderCharacters] {
				return filepath.SkipDir
			}

			return nil
		} else {
			hash := concatLastPathElems(cleanPath, 2)

			if !strings.HasPrefix(info.Name(), hash[FolderCharacters:]) {
				return filepath.SkipDir
			}

			isCommit := HashIsCommit(hash)

			if !isCommit {
				return filepath.SkipDir
			}

			result = hash

			return nil
		}
	})

	if err != nil {
		return "", err
	}

	return result, nil
}
