package maven

import (
	"encoding/json"
	"os"
)

type configFile struct {
	DependencyTree dependencyTree `json:"dependency_tree"`
}

func loadConfiguration(filename string) (*configFile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c configFile
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

type dependencyTree struct {
	ConflictResolution map[string]string `json:"conflict_resolution"`
	Dependencies       []Dep             `json:"dependencies"`
	Version            string            `json:"version"`
}

type Dep struct {
	Coord              string   `json:"coord"`
	Dependencies       []string `json:"dependencies"`
	DirectDependencies []string `json:"directDependencies"`
	File               string   `json:"file"`
	MirrorUrls         []string `json:"mirror_urls,omitempty"`
	Packages           []string `json:"packages"`
	Sha256             string   `json:"sha256,omitempty"`
	URL                string   `json:"url,omitempty"`
	Exclusions         []string `json:"exclusions,omitempty"`
}
