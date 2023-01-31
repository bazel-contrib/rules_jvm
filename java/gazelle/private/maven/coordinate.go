package maven

import (
	"fmt"
	"strings"
)

type coordinate struct {
	GroupID    string
	ArtifactID string
	Type       string
	Classifier string
	Version    string
}

func ParseCoordinate(s string) (*coordinate, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid Maven coordinate %q - needed at least 2 :s", s)
	}
	c := coordinate{
		GroupID:    parts[0],
		ArtifactID: parts[1],
		Version:    parts[len(parts)-1],
	}
	if len(parts) > 3 {
		c.Type = parts[2]
	}
	if len(parts) > 4 {
		c.Classifier = parts[3]
	}
	return &c, nil
}

// ArtifactString returns a value which can be passed to the `artifact()` macro to find a dependency.
func (c *coordinate) ArtifactString() string {
	parts := []string{
		c.GroupID,
		c.ArtifactID,
	}
	if c.Classifier != "" {
		parts = append(parts, c.Classifier)
	}
	return strings.Join(parts, ":")
}
