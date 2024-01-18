package gazelle

import (
	"errors"
	"fmt"
	"sort"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/buildtools/build"
	lru "github.com/hashicorp/golang-lru"
)

const languageName = "java"

// Resolver satisfies the resolve.Resolver interface. It's the
// language-specific resolver extension.
//
// See resolve.Resolver for more information.
type Resolver struct {
	lang          *javaLang
	internalCache *lru.Cache
}

func NewResolver(lang *javaLang) *Resolver {
	internalCache, err := lru.New(10000)
	if err != nil {
		lang.logger.Fatal().Err(err).Msg("error creating cache")
	}

	return &Resolver{
		lang:          lang,
		internalCache: internalCache,
	}
}

func (Resolver) Name() string {
	return languageName
}

func (jr Resolver) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	log := jr.lang.logger.With().Str("step", "Imports").Str("rel", f.Pkg).Str("rule", r.Name()).Logger()

	if !isJavaLibrary(r.Kind()) && r.Kind() != "java_test_suite" {
		return nil
	}

	var out []resolve.ImportSpec
	if pkgs := r.PrivateAttr(packagesKey); pkgs != nil {
		for _, pkg := range pkgs.([]types.ResolvableJavaPackage) {
			out = append(out, resolve.ImportSpec{Lang: languageName, Imp: pkg.String()})
		}
	}

	log.Debug().Str("out", fmt.Sprintf("%#v", out)).Str("label", label.New("", f.Pkg, r.Name()).String()).Msg("return")
	return out
}

func (Resolver) Embeds(r *rule.Rule, from label.Label) []label.Label {
	embedStrings := r.AttrStrings("embed")
	if isJavaProtoLibrary(r.Kind()) {
		embedStrings = append(embedStrings, r.AttrString("proto"))
	}
	embedLabels := make([]label.Label, 0, len(embedStrings))
	for _, s := range embedStrings {
		l, err := label.Parse(s)
		if err != nil {
			continue
		}
		l = l.Abs(from.Repo, from.Pkg)
		embedLabels = append(embedLabels, l)
	}
	return embedLabels
}

func (jr Resolver) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
	resolveInput := imports.(types.ResolveInput)

	packageConfig := c.Exts[languageName].(javaconfig.Configs)[from.Pkg]
	if packageConfig == nil {
		jr.lang.logger.Fatal().Msg("failed retrieving package config")
	}
	isTestRule := isTestRule(r.Kind())

	jr.populateAttr(c, packageConfig, r, "deps", resolveInput.ImportedPackageNames, ix, isTestRule, from, resolveInput.PackageNames)
	jr.populateAttr(c, packageConfig, r, "exports", resolveInput.ExportedPackageNames, ix, isTestRule, from, resolveInput.PackageNames)
}

func (jr Resolver) populateAttr(c *config.Config, pc *javaconfig.Config, r *rule.Rule, attrName string, requiredPackageNames *sorted_set.SortedSet[types.PackageName], ix *resolve.RuleIndex, isTestRule bool, from label.Label, ownPackageNames *sorted_set.SortedSet[types.PackageName]) {
	labels := sorted_set.NewSortedSetFn[label.Label]([]label.Label{}, labelLess)

	for _, implicitDep := range r.AttrStrings(attrName) {
		l, err := label.Parse(implicitDep)
		if err != nil {
			panic(fmt.Sprintf("error converting implicit %s %q to label: %v", attrName, implicitDep, err))
		}
		labels.Add(l)
	}

	for _, imp := range requiredPackageNames.SortedSlice() {
		dep, err := jr.resolveSinglePackage(c, pc, imp, ix, from, isTestRule, ownPackageNames)
		if err != nil {
			jr.lang.logger.Error().Str("import", dep.String()).Err(err).Msg("error converting import")
			panic(fmt.Sprintf("error converting import: %s", err))
		}
		if dep == label.NoLabel {
			continue
		}

		labels.Add(simplifyLabel(c.RepoName, dep, from))
	}

	var exprs []build.Expr
	if labels.Len() > 0 {
		for _, l := range labels.SortedSlice() {
			if l.Relative && l.Name == from.Name {
				continue
			}
			exprs = append(exprs, &build.StringExpr{Value: l.String()})
		}
	}
	if len(exprs) > 0 {
		r.SetAttr(attrName, exprs)
	}
}

func labelLess(l, r label.Label) bool {
	// In UTF-8, / sorts before :
	// We want relative labels to come before absolute ones, so explicitly sort relative before absolute.
	if l.Relative {
		if r.Relative {
			return l.String() < r.String()
		}
		return true
	}
	if r.Relative {
		return false
	}
	return l.String() < r.String()
}

func simplifyLabel(repoName string, l label.Label, from label.Label) label.Label {
	if l.Repo == repoName || l.Repo == "" {
		if l.Pkg == from.Pkg {
			l.Relative = true
		} else {
			l.Repo = ""
		}
	}
	return l
}

func (jr *Resolver) resolveSinglePackage(c *config.Config, pc *javaconfig.Config, imp types.PackageName, ix *resolve.RuleIndex, from label.Label, isTestRule bool, ownPackageNames *sorted_set.SortedSet[types.PackageName]) (out label.Label, err error) {
	cacheKey := types.NewResolvableJavaPackage(imp, false, false)
	importSpec := resolve.ImportSpec{Lang: languageName, Imp: cacheKey.String()}
	if ol, found := resolve.FindRuleWithOverride(c, importSpec, languageName); found {
		return ol, nil
	}

	matches := ix.FindRulesByImportWithConfig(c, importSpec, languageName)
	if len(matches) == 1 {
		return matches[0].Label, nil
	}

	if len(matches) > 1 {
		labels := make([]string, 0, len(matches))
		for _, match := range matches {
			labels = append(labels, match.Label.String())
		}
		sort.Strings(labels)

		jr.lang.logger.Error().
			Str("pkg", imp.Name).
			Strs("targets", labels).
			Msg("resolveSinglePackage found MULTIPLE results in rule index")
	}

	if v, ok := jr.internalCache.Get(cacheKey); ok {
		return simplifyLabel(c.RepoName, v.(label.Label), from), nil
	}

	jr.lang.logger.Debug().Str("parsedImport", imp.Name).Stringer("from", from).Msg("not found yet")

	defer func() {
		if err == nil && out != label.NoLabel {
			jr.internalCache.Add(cacheKey, out)
		}
	}()

	if java.IsStdlib(imp) {
		return label.NoLabel, nil
	}

	// As per https://github.com/bazelbuild/bazel/blob/347407a88fd480fc5e0fbd42cc8196e4356a690b/tools/java/runfiles/Runfiles.java#L41
	if imp.Name == "com.google.devtools.build.runfiles" {
		runfilesLabel := "@bazel_tools//tools/java/runfiles"
		l, err := label.Parse(runfilesLabel)
		if err != nil {
			return label.NoLabel, fmt.Errorf("failed to parse known-good runfiles label %s: %w", runfilesLabel, err)
		}
		return l, nil
	}

	if l, err := jr.lang.mavenResolver.Resolve(imp, pc.ExcludedArtifacts(), pc.MavenRepositoryName()); err != nil {
		var noExternal *maven.NoExternalImportsError
		var multipleExternal *maven.MultipleExternalImportsError

		if errors.As(err, &noExternal) {
			// do not fail, the package might be provided elsewhere
		} else if errors.As(err, &multipleExternal) {
			jr.lang.logger.Error().Msg("Append one of the following to BUILD.bazel:")
			for _, possible := range multipleExternal.PossiblePackages {
				jr.lang.logger.Error().Msgf("# gazelle:resolve java %s %s", imp.Name, possible)
			}
			jr.lang.hasHadErrors = true
		} else {
			jr.lang.logger.Fatal().Err(err).Msg("maven resolver error")
		}
	} else {
		return l, nil
	}

	if isTestRule {
		// If there's exactly one testonly match, use it
		testonlyCacheKey := types.NewResolvableJavaPackage(imp, true, false)
		testonlyImportSpec := resolve.ImportSpec{Lang: languageName, Imp: testonlyCacheKey.String()}
		testonlyMatches := ix.FindRulesByImportWithConfig(c, testonlyImportSpec, languageName)
		if len(testonlyMatches) == 1 {
			cacheKey = testonlyCacheKey
			return simplifyLabel(c.RepoName, testonlyMatches[0].Label, from), nil
		}

		// If there's exactly one testonly match, use it
		testsuiteCacheKey := types.NewResolvableJavaPackage(imp, true, true)
		testsuiteImportSpec := resolve.ImportSpec{Lang: languageName, Imp: testsuiteCacheKey.String()}
		testsuiteMatches := ix.FindRulesByImportWithConfig(c, testsuiteImportSpec, languageName)
		if len(testsuiteMatches) == 1 {
			cacheKey = testsuiteCacheKey
			l := testsuiteMatches[0].Label
			if l != from {
				l.Name += "-test-lib"
				return simplifyLabel(c.RepoName, l, from), nil
			}
		}
	}

	if isTestRule && ownPackageNames.Contains(imp) {
		// Tests may have unique packages which don't exist outside of those tests - don't treat this as an error.
		return label.NoLabel, nil
	}

	jr.lang.logger.Warn().
		Str("package", imp.Name).
		Str("from rule", from.String()).
		Msg("Unable to find package for import in any dependency")

	return label.NoLabel, nil
}

func isJavaLibrary(kind string) bool {
	return kind == "java_library" || isJavaProtoLibrary(kind)
}

func isJavaProtoLibrary(kind string) bool {
	return kind == "java_proto_library" || kind == "java_grpc_library"
}
