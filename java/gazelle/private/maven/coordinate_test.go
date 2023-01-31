package maven

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	type testCase struct {
		wantCoordinate *coordinate
		wantErr        error
	}
	for coord, want := range map[string]testCase{
		"com.google.guava:guava:31.1-jre": {
			wantCoordinate: &coordinate{
				GroupID:    "com.google.guava",
				ArtifactID: "guava",
				Version:    "31.1-jre",
			},
		},
		"com.google.guava:guava:jar:sources:31.1-jre": {
			wantCoordinate: &coordinate{
				GroupID:    "com.google.guava",
				ArtifactID: "guava",
				Type:       "jar",
				Classifier: "sources",
				Version:    "31.1-jre",
			},
		},
		"com.google.guava:guava": {
			wantErr: fmt.Errorf("invalid Maven coordinate \"com.google.guava:guava\" - needed at least 2 :s"),
		},
	} {
		t.Run(coord, func(t *testing.T) {
			got, err := ParseCoordinate(coord)
			require.Equal(t, want.wantErr, err)
			require.Equal(t, want.wantCoordinate, got)
		})
	}
}
