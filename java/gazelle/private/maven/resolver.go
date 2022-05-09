package maven

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/cmd/parsejars/manifest"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/bazel"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven/multiset"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/tools_jvm_autodeps/listclassesinjar"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Resolver interface {
	Resolve(pkg string) (label.Label, error)
}

type Manifester interface {
	DumpManifest() *manifest.Manifest
}

// resolver finds Maven provided packages by reading the maven_install.json
// file from rules_jvm_external, and by listing classes in the JAR files.
type resolver struct {
	data   *multiset.StringMultiSet
	logger zerolog.Logger
}

func NewResolverFromManifest(manifestFile, installFile string, logger zerolog.Logger) (Resolver, error) {
	var f manifest.File
	if err := f.Decode(manifestFile); err != nil {
		return nil, fmt.Errorf("manifest error: %w", err)
	}

	ok, err := f.VerifyIntegrity(installFile)
	if err != nil {
		return nil, fmt.Errorf("verify error: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("not up-to-date manifest")
	}

	r := resolver{
		data:   multiset.NewStringMultiSet(),
		logger: logger.With().Str("_c", "maven-resolver").Logger(),
	}

	for k, vs := range f.Manifest.ArtifactsMapping {
		for _, v := range vs {
			r.data.Add(k, v)
		}
	}

	return &r, nil
}

func NewResolver(outputBase, installFile string, logger zerolog.Logger) (Resolver, error) {
	r := resolver{
		data:   multiset.NewStringMultiSet(),
		logger: logger.With().Str("_c", "maven-resolver").Logger(),
	}

	c, err := loadConfiguration(installFile)
	if err != nil {
		r.logger.Warn().Err(err).Msg("not loading maven dependencies")
		return &r, nil
	}

	r.logger.Debug().Int("count", len(c.DependencyTree.Dependencies)).Msg("Dependency count")
	r.logger.Debug().Str("conflicts", fmt.Sprintf("%#v", c.DependencyTree.ConflictResolution)).Msg("Maven install conflict")

	loadStart := time.Now()

	var g errgroup.Group
	for _, dep := range c.DependencyTree.Dependencies {
		dep := dep
		g.Go(func() error {
			// FIXME this might change with https://github.com/bazelbuild/rules_jvm_external/issues/473
			f := filepath.Join(outputBase, "external", bazel.CleanupLabel(dep.Coord), "file", "downloaded")
			if !fileExists(f) {
				if strings.Contains(f, "_jar_sources_") {
					return nil
				}

				r.logger.Warn().Str("artifact", dep.Coord).Str("f", f).Msgf("missing artifact %s: run `bazel fetch @maven//...` first", dep.Coord)
				return nil
			}

			cls, err := listclassesinjar.List(f)
			if err != nil {
				return fmt.Errorf("error listing classes from %s: %s", f, err)
			}

			for _, c := range cls {
				pkg := java.NewImport(string(c)).Pkg
				l := label.New("maven", "", bazel.CleanupLabel(artifactFromCoord(dep.Coord)))
				r.data.Add(pkg, l.String())
			}

			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	r.logger.Debug().Str("duration", time.Since(loadStart).String()).Msg("Loaded classes from JAR files")

	return &r, nil
}

func (r *resolver) Resolve(pkg string) (label.Label, error) {
	v, found := r.data.Get(pkg)
	if !found {
		return label.NoLabel, fmt.Errorf("package not found: %s", pkg)
	}

	switch len(v) {
	case 0:
		return label.NoLabel, errors.New("no external imports")

	case 1:
		var ret string
		for r := range v {
			ret = r
			break
		}
		return label.Parse(ret)

	default:
		r.logger.Error().Msg("Append one of the following to BUILD.bazel:")
		for k := range v {
			r.logger.Error().Msgf("# gazelle:resolve java %s %s", pkg, k)
		}

		return label.NoLabel, errors.New("many possible imports")
	}
}

func (r *resolver) DumpManifest() *manifest.Manifest {
	return r.data.DumpManifest()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func artifactFromCoord(coord string) string {
	g, a, _ := splitCoord(coord)
	return strings.Join([]string{g, a}, ":")
}

func splitCoord(coord string) (groupId, artifactId, version string) {
	parts := strings.Split(coord, ":")
	return parts[0], parts[1], parts[len(parts)-1]
}

func LabelFromArtifact(artifact string) string {
	return label.New("maven", "", bazel.CleanupLabel(artifact)).String()
}
