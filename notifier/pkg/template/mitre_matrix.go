package template

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const mitreMatrixFileName = "mitre_matrix.json"

type mitreMatrixData struct {
	MatrixVersion string        `json:"matrix_version"`
	Tactics       []mitreTactic `json:"tactics"`
}

type mitreTactic struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Techniques []mitreTechnique `json:"techniques"`
}

type mitreTechnique struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	tacticsByID    map[string]string
	techniquesByID map[string]string
)

func initMitreMatrixData(templatesFolder string) {
	filePath := filepath.Join(templatesFolder, mitreMatrixFileName)

	raw, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var data mitreMatrixData
	if err := json.Unmarshal(raw, &data); err != nil {
		panic(err)
	}

	newTactics := make(map[string]string, len(data.Tactics))
	newTechniques := make(map[string]string)
	for _, tactic := range data.Tactics {
		newTactics[tactic.ID] = tactic.Name
		for _, technique := range tactic.Techniques {
			newTechniques[technique.ID] = technique.Name
		}
	}

	tacticsByID = newTactics
	techniquesByID = newTechniques
}

func tacticName(id string) string {
	if name, ok := tacticsByID[id]; ok {
		return name
	}
	return id
}

func techniqueName(id string) string {
	if name, ok := techniquesByID[id]; ok {
		return name
	}
	return id
}
