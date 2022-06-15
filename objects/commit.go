package objects

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"lit/util"
	"os"
	"time"
)

type Commit struct {
	Name, CommitTree string
	Parents          []string
	Time             time.Time
}

func (*Commit) ObjectType() string {
	return "commit"
}

func NewCommit(name, commitTree string, time time.Time) *Commit {
	return &Commit{name, commitTree, []string{}, time}
}

func HashIsCommit(hash string) bool {
	_, err := ReadAsCommit(hash)

	return err == nil
}

func HashCommit(commit *Commit) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", commit)))

	return fmt.Sprintf("%2x", hash)
}

func CommitToJSON(commit *Commit) []byte {
	data, err := json.MarshalIndent(map[string]any{
		"Type": "Commit",
		"Object": struct {
			Com Commit `json:"Commit"`
		}{*commit},
	}, "", "\t")

	if err != nil {
		panic(err)
	}

	return data
}

func WriteCommit(commit *Commit) (hash string) {
	hash = HashCommit(commit)

	jsonCommit := CommitToJSON(commit)

	err := os.MkdirAll(".lit/objects/"+hash[:FolderCharacters], 0777)

	if err != nil {
		hash = ""
		return
	}

	err = os.WriteFile(GetHashPath(hash), jsonCommit, 0777)

	if err != nil {
		hash = ""
	}

	return
}

func ReadAsCommit(hash string) (*Commit, error) {
	objectData := genericObject{}

	err := util.ReadJSON(GetHashPath(hash), &objectData)

	if err != nil {
		return nil, ErrCouldNotRead
	}

	if objectData.ObjType != "Commit" {
		return nil, ErrNotOfType
	}

	commitStructure := objectData.Obj["Commit"].(map[string]any)

	resultTime := time.Time{}

	err = json.Unmarshal([]byte(`"`+commitStructure["Time"].(string)+`"`), &resultTime)

	if err != nil {
		panic(err)
	}

	c := NewCommit(commitStructure["Name"].(string), commitStructure["CommitTree"].(string), resultTime)

	for _, parent := range commitStructure["Parents"].([]any) {
		c.Parents = append(c.Parents, parent.(string))
	}

	return c, nil
}
