package util

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"
)

func CleanPath(path string) string {
	return pathlib.Clean(filepath.ToSlash(path))
}

func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)

	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

func WriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "\t")
	
	if err != nil {
		panic(err)
	}

	dir, _ := filepath.Split(path)
	err = os.MkdirAll(dir, 0777)

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0777)
}

func ReadJSON(path string, v any) error {
	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v)

	if err != nil {
		return err
	}

	return nil
}

var (
	ErrFileWalk = errors.New("error while walking files")
)

func ForeachSubfile(dir string, f func(string, fs.DirEntry) error) error {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		cleanPath := CleanPath(path)
		if d.IsDir() {
			return nil
		}

		if err != nil {
			return ErrFileWalk
		}

		return f(cleanPath, d)
	})

	return err
}

func IsSubPath(base string, sub string) bool {
	relative, err := filepath.Rel(filepath.FromSlash(base), filepath.FromSlash(sub))

	if err != nil {
		return false
	}

	return !strings.Contains(pathlib.Clean(relative), "..")
}

