package objects

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"lit/util"
	"os"
)

type Blob struct {
	Content string
}

func (*Blob) ObjectType() string {
	return "blob"
}

func NewBlob(content string) *Blob {
	return &Blob{content}
}

func Hash(data []byte) string {
	return fmt.Sprintf("%2x", sha256.Sum256(data))
}

func BlobToJSON(data []byte) []byte {
	data, err := json.MarshalIndent(map[string]any{
		"Type": "Blob",
		"Object": struct {
			Content string
		}{string(data)},
	}, "", "\t")

	if err != nil {
		panic(err)
	}

	return data
}

func Blobify(path string) string {
	data, err := os.ReadFile(path)

	if err != nil {
		return ""
	}

	hash := WriteBlob(data)

	return hash
}

func GetHashPath(hash string) string {
	return ".lit/objects/" + hash[:FolderCharacters] + "/" + hash[FolderCharacters:]
}

func WriteBlob(data []byte) (hash string) {
	hash = Hash(data)

	jsonBlob := BlobToJSON(data)

	err := os.MkdirAll(".lit/objects/"+hash[:FolderCharacters], 0777)

	if err != nil {
		hash = ""
		return
	}

	err = os.WriteFile(GetHashPath(hash), jsonBlob, 0777)

	if err != nil {
		hash = ""
	}

	return
}

type genericObject struct {
	ObjType string         `json:"Type"`
	Obj     map[string]any `json:"Object"`
}

func ReadAsBlob(hash string) (string, error) {
	objectData := genericObject{}

	err := util.ReadJSON(GetHashPath(hash), &objectData)

	if err != nil {
		return "", ErrCouldNotRead
	}

	if objectData.ObjType != "Blob" {
		return "", ErrNotOfType
	}

	return objectData.Obj["Content"].(string), nil
}
