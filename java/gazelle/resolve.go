package gazelle

import (
	"errors"
	"fmt"
	"sort"
	"strings"

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

func (*Resolver) Name() string {
	return languageName
}

func (jr *Resolver) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	log := jr.lang.logger.With().Str("step", "Imports").Str("rel", f.Pkg).Str("rule", r.Name()).Logger()

	if !isJavaLibrary(r.Kind()) && r.Kind() != "java_test_suite" && r.Kind() != "java_export" {
		return nil
	}

	lbl := label.New("", f.Pkg, r.Name())
	deps := attrLabels("deps", r, f)

	if r.Kind() == "java_export" {
		var out []resolve.ImportSpec
		for _, dep := range deps {
			// Because Imports is a post-order traversal, the children will be added before the parent, so we always have to take the rightmost element of this package to get the closest one to the root of the tree.
			jr.lang.javaExportCache[dep] = append(jr.lang.javaExportCache[dep], lbl)
		}

		labelsToVisit := make([]label.Label, len(deps))
		_ = copy(labelsToVisit, deps)

		queuedForVisit := make(map[label.Label]bool)
		for _, depLabel := range labelsToVisit {
			queuedForVisit[depLabel] = true
		}

		for len(labelsToVisit) > 0 {
			dep := labelsToVisit[0]
			labelsToVisit = labelsToVisit[1:]

			resolveInputForDep := jr.lang.labelsToResolveInputs[dep]
			for _, pkg := range resolveInputForDep.PackageNames.SortedSlice() {
				out = append(out, resolve.ImportSpec{
					Lang: languageName, Imp: pkg.Name,
				})
			}

			for _, importedPkg := range resolveInputForDep.ImportedPackageNames.SortedSlice() {
				lblToVisit := jr.lang.packagesToLabelsDeclaringThem[importedPkg]
				if _, labelInAnotherJavaExport := jr.lang.labelToJavaExport[lblToVisit]; labelInAnotherJavaExport {
					continue
				}
				if ok := queuedForVisit[lblToVisit]; !ok {
					labelsToVisit = append(labelsToVisit, lblToVisit)
					queuedForVisit[lblToVisit] = true
				}
			}
		}

		jr.lang.javaExportsTransitiveDeps[lbl] = queuedForVisit
		for queued, _ := range queuedForVisit {
			//if _, ok := jr.lang.labelToJavaExport[queued]; ok {
			//	panic("Target in multiple exports: %s")
			//}
			jr.lang.labelToJavaExport[queued] = lbl
		}

		return out
	}

	var out []resolve.ImportSpec
	if pkgs := r.PrivateAttr(packagesKey); pkgs != nil {
		for _, pkg := range pkgs.([]types.ResolvableJavaPackage) {
			out = append(out, resolve.ImportSpec{Lang: languageName, Imp: pkg.String()})
		}
	}

	log.Debug().Str("out", fmt.Sprintf("%#v", out)).Str("label", lbl.String()).Msg("return")
	return out
}

func attrLabels(attr string, r *rule.Rule, f *rule.File) []label.Label {
	depsStrings := r.AttrStrings(attr)
	deps := make([]label.Label, 0, len(depsStrings))
	for _, depString := range depsStrings {
		lbl, err := label.Parse(depString)
		if err != nil {
			// TODO BL: handle
		}
		if lbl.Pkg == "" {
			lbl.Pkg = f.Pkg
			lbl.Relative = false
		}
		deps = append(deps, lbl)
	}
	return deps
}

func (*Resolver) Embeds(r *rule.Rule, from label.Label) []label.Label {
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

func (jr *Resolver) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
	resolveInput := imports.(types.ResolveInput)

	packageConfig := c.Exts[languageName].(javaconfig.Configs)[from.Pkg]
	if packageConfig == nil {
		jr.lang.logger.Fatal().Msg("failed retrieving package config")
	}
	isTestRule := packageConfig.IsTestRule(r.Kind())
	if literalExpr, ok := r.Attr("testonly").(*build.LiteralExpr); ok {
		if literalExpr.Token == "True" {
			isTestRule = true
		}
	}

	// If the current library is exported under a `java_export`, it shouldn't be visible for targets outside of the java_export.
	if containingJavaExport, inJavaExport := jr.lang.labelToJavaExport[from]; inJavaExport {
		r.SetAttr("visibility", []string{fmt.Sprintf("//%s:__pkg__", containingJavaExport.Pkg)})
	}

	jr.populateAttr(c, packageConfig, r, "deps", resolveInput.ImportedPackageNames, ix, isTestRule, from, resolveInput.PackageNames)
	jr.populateAttr(c, packageConfig, r, "exports", resolveInput.ExportedPackageNames, ix, isTestRule, from, resolveInput.PackageNames)

	jr.populatePluginsAttr(c, ix, resolveInput, packageConfig, from, isTestRule, r)
}

func (jr *Resolver) populateAttr(c *config.Config, pc *javaconfig.Config, r *rule.Rule, attrName string, requiredPackageNames *sorted_set.SortedSet[types.PackageName], ix *resolve.RuleIndex, isTestRule bool, from label.Label, ownPackageNames *sorted_set.SortedSet[types.PackageName]) {
	labels := sorted_set.NewSortedSetFn[label.Label]([]label.Label{}, labelLess)

	for _, imp := range requiredPackageNames.SortedSlice() {
		dep := jr.resolveSinglePackage(c, pc, imp, ix, from, isTestRule, ownPackageNames)
		if dep == label.NoLabel {
			continue
		}

		javaExportsWithPackage := jr.lang.javaExportCache[dep]
		if len(javaExportsWithPackage) != 0 {
			javaExportDep := javaExportsWithPackage[len(javaExportsWithPackage)-1]
			dep = javaExportDep
		}

		labels.Add(simplifyLabel(c.RepoName, dep, from))
	}

	setLabelAttrIncludingExistingValues(r, attrName, labels)

}

func (jr *Resolver) populatePluginsAttr(c *config.Config, ix *resolve.RuleIndex, resolveInput types.ResolveInput, packageConfig *javaconfig.Config, from label.Label, isTestRule bool, r *rule.Rule) {
	pluginLabels := sorted_set.NewSortedSetFn[label.Label]([]label.Label{}, labelLess)
	for _, annotationProcessor := range resolveInput.AnnotationProcessors.SortedSlice() {
		dep := jr.resolveSinglePackage(c, packageConfig, annotationProcessor.PackageName(), ix, from, isTestRule, resolveInput.PackageNames)
		if dep == label.NoLabel {
			continue
		}

		// Use the naming scheme for plugins as per https://github.com/bazelbuild/rules_jvm_external/pull/1102
		// In the case of overrides (i.e. # gazelle:resolve targets) we require that they follow the same name-mangling scheme for the java_plugin target as rules_jvm_external uses.
		// Ideally this would be a call to `java_plugin_artifact(dep.String(), annotationProcessor.FullyQualifiedClassName())` but we don't have function calls working in attributes.
		dep.Name += "__java_plugin__" + strings.NewReplacer(".", "_", "$", "_").Replace(annotationProcessor.FullyQualifiedClassName())

		pluginLabels.Add(simplifyLabel(c.RepoName, dep, from))
	}

	setLabelAttrIncludingExistingValues(r, "plugins", pluginLabels)
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

// Note: This function may modify labels.
func setLabelAttrIncludingExistingValues(r *rule.Rule, attrName string, labels *sorted_set.SortedSet[label.Label]) {
	for _, implicitDep := range r.AttrStrings(attrName) {
		l, err := label.Parse(implicitDep)
		if err != nil {
			panic(fmt.Sprintf("error converting implicit %s %q to label: %v", attrName, implicitDep, err))
		}
		labels.Add(l)
	}

	var exprs []build.Expr
	if labels.Len() > 0 {
		for _, l := range labels.SortedSlice() {
			if l.Relative && l.Name == r.Name() {
				continue
			}
			exprs = append(exprs, &build.StringExpr{Value: l.String()})
		}
	}
	if len(exprs) > 0 {
		r.SetAttr(attrName, exprs)
	}
}

func (jr *Resolver) resolveSinglePackage(c *config.Config, pc *javaconfig.Config, imp types.PackageName, ix *resolve.RuleIndex, from label.Label, isTestRule bool, ownPackageNames *sorted_set.SortedSet[types.PackageName]) (out label.Label) {
	cacheKey := types.NewResolvableJavaPackage(imp, false, false)
	importSpec := resolve.ImportSpec{Lang: languageName, Imp: cacheKey.String()}
	if ol, found := resolve.FindRuleWithOverride(c, importSpec, languageName); found {
		return ol
	}

	matches := ix.FindRulesByImportWithConfig(c, importSpec, languageName)
	if len(matches) == 1 {
		return matches[0].Label
	}

	if len(matches) > 1 {
		matches = jr.tryResolvingToJavaExport(matches, from)
	}

	if len(matches) == 1 {
		return matches[0].Label
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
		return simplifyLabel(c.RepoName, v.(label.Label), from)
	}

	jr.lang.logger.Debug().Str("parsedImport", imp.Name).Stringer("from", from).Msg("not found yet")

	defer func() {
		if out != label.NoLabel {
			jr.internalCache.Add(cacheKey, out)
		}
	}()

	if java.IsStdlib(imp) {
		return label.NoLabel
	}

	// As per https://github.com/bazelbuild/bazel/blob/347407a88fd480fc5e0fbd42cc8196e4356a690b/tools/java/runfiles/Runfiles.java#L41
	if imp.Name == "com.google.devtools.build.runfiles" {
		runfilesLabel := "@bazel_tools//tools/java/runfiles"
		l, err := label.Parse(runfilesLabel)
		if err != nil {
			jr.lang.logger.Fatal().Str("label", runfilesLabel).Err(err).Msg("failed to parse known-good runfiles label")
			return label.NoLabel
		}
		return l
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
		return l
	}

	if isTestRule {
		// If there's exactly one testonly match, use it
		testonlyCacheKey := types.NewResolvableJavaPackage(imp, true, false)
		testonlyImportSpec := resolve.ImportSpec{Lang: languageName, Imp: testonlyCacheKey.String()}
		testonlyMatches := ix.FindRulesByImportWithConfig(c, testonlyImportSpec, languageName)
		if len(testonlyMatches) == 1 {
			cacheKey = testonlyCacheKey
			return simplifyLabel(c.RepoName, testonlyMatches[0].Label, from)
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
				return simplifyLabel(c.RepoName, l, from)
			}
		}
	}

	if isTestRule && ownPackageNames.Contains(imp) {
		// Tests may have unique packages which don't exist outside of those tests - don't treat this as an error.
		return label.NoLabel
	}

	jr.lang.logger.Warn().
		Str("package", imp.Name).
		Str("from rule", from.String()).
		Msg("Unable to find package for import in any dependency")
	jr.lang.hasHadErrors = true

	return label.NoLabel
}

// tryResolvingToJavaExport attempts to narrow down a list of resolution candidates by preferring java_export targets when appropriate.
// A dependency will be resolved to a `java_export` target when the following are all true.
//   - The dependency is contained in a java_export target, and
//   - There is exactly one java_export target that contains the dependency, and
//   - That java_export does not export the target under consideration (`from`).
//
// Returns a subset of `results`, either by picking an appropriate `java_export`, or by eliminating ineligible `java_export`.
// The program will issue a fatal error if it finds that more than one java_export contains the required dependency.
func (jr *Resolver) tryResolvingToJavaExport(results []resolve.FindResult, from label.Label) []resolve.FindResult {
	var javaExportsThatCoverThisDep []resolve.FindResult
	var nonJavaExportResults []resolve.FindResult
	for _, result := range results {
		_, depIsJavaExport := jr.lang.javaExportsTransitiveDeps[result.Label]
		if depIsJavaExport {
			javaExportsThatCoverThisDep = append(javaExportsThatCoverThisDep, result)
		} else {
			nonJavaExportResults = append(nonJavaExportResults, result)
		}
	}

	if len(javaExportsThatCoverThisDep) == 0 {
		return results
	} else if len(javaExportsThatCoverThisDep) > 1 {
		var exportStrings []string
		for _, exportResult := range javaExportsThatCoverThisDep {
			exportStrings = append(exportStrings, exportResult.Label.String())
		}
		jr.lang.logger.Fatal().
			Str("rule", from.Pkg).
			Strs("java_exports", exportStrings).
			Msg("resolveSinglePackage found MULTIPLE java_export targets exporting this rule")
	}

	javaExportForDep := javaExportsThatCoverThisDep[0]
	myJavaExport := jr.lang.labelToJavaExport[from]
	// If we're inside the same java export (or inside no java export), return the dep
	if myJavaExport != javaExportForDep.Label {
		return []resolve.FindResult{javaExportForDep}
	}

	// If we don't find any relevant java_export, resolve normally.
	return nonJavaExportResults
}

func isJavaLibrary(kind string) bool {
	return kind == "java_library" || isJavaProtoLibrary(kind)
}

func isJavaProtoLibrary(kind string) bool {
	return kind == "java_proto_library" || kind == "java_grpc_library"
}
