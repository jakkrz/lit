package objects

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"lit/util"
	"os"
)

type TreeEntry struct {
	ObjType string `json:"Type"`
	Hash string
}

func HashTree(entries map[string]TreeEntry) string {
	toHash := ""
	for name, entry := range entries {
		toHash += name + entry.ObjType + entry.Hash + "\n"
	}

	return fmt.Sprintf("%2x", sha256.Sum256([]byte(toHash)))
}

func TreeToJSON(entries map[string]TreeEntry) []byte {
	data, err := json.MarshalIndent(map[string]any{
		"Type": "Tree",
		"Object": struct {
			Entries map[string]TreeEntry
		}{entries},
	}, "", "\t")

	if err != nil {
		panic(err)
	}

	return data
}

func WriteTree(entries map[string]TreeEntry) (hash string) {
	hash = HashTree(entries)

	jsonTree := TreeToJSON(entries)

	err := os.MkdirAll(".lit/objects/"+hash[:FolderCharacters], 0777)

	if err != nil {
		hash = ""
		return
	}

	err = os.WriteFile(GetHashPath(hash), jsonTree, 0777)

	if err != nil {
		hash = ""
	}

	return
}

func ReadAsTree(hash string) (map[string]TreeEntry, error) {
	objectData := genericObject{}

	err := util.ReadJSON(GetHashPath(hash), &objectData)

	if err != nil {
		return nil, ErrCouldNotRead
	}

	if objectData.ObjType != "Tree" {
		return nil, ErrNotOfType
	}

	entries := objectData.Obj["Entries"].(map[string]any)
	result := map[string]TreeEntry{}

	for name, entry := range entries {
		entryAsMap := entry.(map[string]any)
		result[name] = TreeEntry{Hash: entryAsMap["Hash"].(string), ObjType: entryAsMap["Type"].(string)}
	}

	return result, nil
}
