package maven

import (
	"errors"
	"fmt"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/bazel"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven/multiset"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/rs/zerolog"
)

type Resolver interface {
	Resolve(pkg types.PackageName, excludedArtifacts map[string]struct{}, mavenRepositoryName string) (label.Label, error)
}

// resolver finds Maven provided packages by reading the maven_install.json
// file from rules_jvm_external.
type resolver struct {
	data   *multiset.StringMultiSet
	logger zerolog.Logger
}

func NewResolver(installFile string, logger zerolog.Logger) (Resolver, error) {
	r := resolver{
		data:   multiset.NewStringMultiSet(),
		logger: logger.With().Str("_c", "maven-resolver").Logger(),
	}

	c, err := loadConfiguration(installFile)
	if err != nil {
		r.logger.Warn().Err(err).Msg("not loading maven dependencies")
		return &r, nil
	}

	dependencies := c.ListDependencies()

	r.logger.Debug().Int("count", len(dependencies)).Msg("Dependency count")

	for _, depName := range dependencies {
		coords, err := ParseCoordinate(c.GetDependencyCoordinates(depName))
		if err != nil {
			return nil, fmt.Errorf("failed to parse coordinate %v: %w", coords, err)
		}
		for _, pkg := range c.ListDependencyPackages(depName) {
			r.data.Add(pkg, coords.ArtifactString())
		}
	}

	return &r, nil
}

func (r *resolver) Resolve(pkg types.PackageName, excludedArtifacts map[string]struct{}, mavenRepositoryName string) (label.Label, error) {
	v, found := r.data.Get(pkg.Name)
	if !found {
		return label.NoLabel, fmt.Errorf("package not found: %s", pkg)
	}

	var filtered []string
	for k := range v {
		if _, excluded := excludedArtifacts[LabelFromArtifact(mavenRepositoryName, k).String()]; excluded {
			continue
		}
		filtered = append(filtered, LabelFromArtifact(mavenRepositoryName, k).String())
	}

	switch len(filtered) {
	case 0:
		return label.NoLabel, errors.New("no external imports")

	case 1:
		var ret string
		for _, r := range filtered {
			ret = r
			break
		}
		return label.Parse(ret)

	default:
		r.logger.Error().Msg("Append one of the following to BUILD.bazel:")
		for _, k := range filtered {
			r.logger.Error().Msgf("# gazelle:resolve java %s %s", pkg.Name, k)
		}

		return label.NoLabel, errors.New("many possible imports")
	}
}

func LabelFromArtifact(mavenRepositoryName string, artifact string) label.Label {
	return label.New(mavenRepositoryName, "", bazel.CleanupLabel(artifact))
}
