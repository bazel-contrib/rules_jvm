package gazelle

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/kotlin"
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
	// classIndex is a lazy per-package index, built only for packages with ambiguous
	// resolution (split packages). Maintains prod/test distinction.
	classIndex map[types.PackageName]*packageClassIndex
	// configs provides a map from pkg to config. This allows us to use the config in
	// Embeds.
	configs map[string]*config.Config
}

// packageClassIndex maps class names to their providing labels for a single package.
// Built lazily per-package only when that package has ambiguous resolution.
type packageClassIndex struct {
	// prod maps bare outer class name -> providers (non-testonly rules)
	prod map[string][]label.Label
	// test maps bare outer class name -> providers (testonly rules)
	test map[string][]label.Label
}

func NewResolver(lang *javaLang) *Resolver {
	internalCache, err := lru.New(10000)
	if err != nil {
		lang.logger.Fatal().Err(err).Msg("error creating cache")
	}

	return &Resolver{
		lang:          lang,
		internalCache: internalCache,
		classIndex:    make(map[types.PackageName]*packageClassIndex),
		configs:       make(map[string]*config.Config),
	}
}

func (*Resolver) Name() string {
	return languageName
}

func (jr *Resolver) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	log := jr.lang.logger.With().Str("step", "Imports").Str("rel", f.Pkg).Str("rule", r.Name()).Logger()

	// Cache config for use in Embeds, which doesn't receive config in its interface
	jr.configs[f.Pkg] = c

	if !isJvmLibrary(c, r.Kind()) && r.Kind() != "java_test_suite" && r.Kind() != "java_export" {
		return nil
	}

	lbl := label.New("", f.Pkg, r.Name())

	var out []resolve.ImportSpec
	if pkgs := r.PrivateAttr(packagesKey); pkgs != nil {
		for _, pkg := range pkgs.([]types.ResolvableJavaPackage) {
			out = append(out, resolve.ImportSpec{Lang: languageName, Imp: pkg.String()})
		}
	}
	// NOTE: We intentionally do NOT register classes in Gazelle's global RuleIndex.
	// Class-level resolution uses a lazy, per-package index built only when needed
	// (when package-level resolution is ambiguous due to split packages).
	// This keeps the global index small and fast.

	log.Debug().Str("out", fmt.Sprintf("%#v", out)).Str("label", lbl.String()).Msg("return")
	return out
}

func (jr *Resolver) Embeds(r *rule.Rule, from label.Label) []label.Label {
	embedStrings := r.AttrStrings("embed")
	if isJavaProtoLibrary(jr.configs[from.Pkg], r.Kind()) {
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

	// If the current library is exported under a `java_export`, it shouldn't be visible for targets outside the java_export.
	if packageConfig.ResolveToJavaExports() && isJavaLibrary(c, r.Kind()) {
		visibility := jr.lang.javaExportIndex.VisibilityForLabel(from)
		if visibility != nil {
			var asStrings []string
			for _, vis := range visibility.SortedSlice() {
				asStrings = append(asStrings, vis.String())
			}
			// The rule attr replacement code is buggy, because while in `rule.SetAttr` we can replace the RHS of the expression, attr.val is always unchanged. I suspect it has to do with pointer magic.
			// Fixed in https://github.com/bazel-contrib/bazel-gazelle/issues/2045
			r.DelAttr("visibility")
			r.SetAttr("visibility", asStrings)
		}
	}

	jr.populateAttr(c, packageConfig, r, "deps", resolveInput.ImportedPackageNames, resolveInput.ImportedClasses, ix, isTestRule, from, resolveInput.PackageNames)
	jr.populateAttr(c, packageConfig, r, "exports", resolveInput.ExportedPackageNames, resolveInput.ExportedClassNames, ix, isTestRule, from, resolveInput.PackageNames)

	jr.populatePluginsAttr(c, ix, resolveInput, packageConfig, from, isTestRule, r)
}

func (jr *Resolver) populateAttr(c *config.Config, pc *javaconfig.Config, r *rule.Rule, attrName string, requiredPackageNames *sorted_set.SortedSet[types.PackageName], importedClasses *sorted_set.SortedSet[types.ClassName], ix *resolve.RuleIndex, isTestRule bool, from label.Label, ownPackageNames *sorted_set.SortedSet[types.PackageName]) {
	labels := sorted_set.NewSortedSetFn[label.Label]([]label.Label{}, sorted_set.LabelLess)

	// Build a map of package -> classes for efficient lookup during class-level resolution
	classesByPackage := make(map[types.PackageName][]types.ClassName)
	if importedClasses != nil {
		for _, cls := range importedClasses.SortedSlice() {
			pkg := cls.PackageName()
			classesByPackage[pkg] = append(classesByPackage[pkg], cls)
		}
	}

	for _, imp := range requiredPackageNames.SortedSlice() {
		var pkgClasses []string
		for _, cls := range classesByPackage[imp] {
			pkgClasses = append(pkgClasses, cls.BareOuterClassName())
		}

		// Check if any imported class has an explicit resolve directive.
		// If so, we must use class-level resolution to respect those overrides,
		// since package-level resolution (including resolve_regexp) would otherwise
		// take precedence and ignore class-specific directives.
		hasClassOverrides := false
		if len(classesByPackage[imp]) > 0 {
			for _, className := range classesByPackage[imp] {
				classImportSpec := resolve.ImportSpec{Lang: languageName, Imp: className.FullyQualifiedClassName()}
				if _, found := resolve.FindRuleWithOverride(c, classImportSpec, languageName); found {
					hasClassOverrides = true
					break
				}
			}
		}

		// If there are class-level overrides, skip package-level resolution and go
		// directly to class-level resolution to ensure overrides are respected.
		if hasClassOverrides {
			jr.lang.logger.Debug().
				Str("package", imp.Name).
				Strs("classes", pkgClasses).
				Stringer("from", from).
				Msg("class-level resolve directive found, using class-level resolution")

			for _, className := range classesByPackage[imp] {
				// Check for explicit resolve directive for this specific class first
				classImportSpec := resolve.ImportSpec{Lang: languageName, Imp: className.FullyQualifiedClassName()}
				if ol, found := resolve.FindRuleWithOverride(c, classImportSpec, languageName); found {
					labels.Add(simplifyLabel(c.RepoName, ol, from))
					continue
				}

				l, err := jr.lang.mavenResolver.ResolveClass(className, pc.ExcludedArtifacts(), pc.MavenRepositoryName())
				if err != nil {
					jr.lang.logger.Warn().Err(err).Str("class", className.FullyQualifiedClassName()).Msg("error resolving class")
					continue
				}
				if l == label.NoLabel {
					l = jr.resolveSingleClass(c, pc, className, ix, from, isTestRule)
				}
				if l != label.NoLabel {
					labels.Add(simplifyLabel(c.RepoName, l, from))
				}
			}
			continue
		}

		// Try package-level resolution first (fast path)
		dep, ambiguous := jr.resolveSinglePackageWithAmbiguity(c, pc, imp, ix, from, isTestRule, ownPackageNames, pkgClasses)
		if dep != label.NoLabel {
			labels.Add(simplifyLabel(c.RepoName, dep, from))

			// The package resolved unambiguously to a single target, but an external
			// gazelle plugin (e.g. a proto/wire generator) may own some classes of the same
			// package. Class import specs are never registered in the global index, so a
			// class-level lookup always falls through to registered CrossResolvers. Only
			// probe classes the resolved target doesn't already declare, to keep the fast
			// path fast when there is no external provider.
			for _, className := range classesByPackage[imp] {
				if jr.ruleDeclaresClass(dep, className) {
					continue
				}
				if l := jr.resolveClassFromCrossResolver(c, pc, className, ix, from); l != label.NoLabel {
					labels.Add(l)
					continue
				}
				if isTestRule {
					// A test may import a class from a testonly library (e.g. a testFixtures source
					// set) whose package the resolved production target also owns; the in-repo class
					// index includes testonly providers for test rules.
					if l := jr.resolveSingleClass(c, pc, className, ix, from, true); l != label.NoLabel {
						labels.Add(l)
						continue
					}
					// Or the class may be a helper in another package's java_test_suite (its
					// "<suite>-test-lib"). Depend on that helper library, but only when it actually
					// declares the class -- a class the production provider doesn't declare may be a
					// main top-level function, not a test helper, which must not pull the lib in.
					if l := jr.resolveTestSuiteHelperClass(c, imp, className, ix, from); l != label.NoLabel {
						labels.Add(l)
						continue
					}
				}
			}
			continue
		}

		// Only fall back to class-level resolution when package resolution is ambiguous
		if ambiguous && len(classesByPackage[imp]) > 0 {
			jr.lang.logger.Debug().
				Str("package", imp.Name).
				Strs("classes", pkgClasses).
				Stringer("from", from).
				Msg("package has multiple providers, attempting class-level resolution")

			resolvedAny := false
			for _, className := range classesByPackage[imp] {
				// Check for explicit resolve directive for this specific class first
				classImportSpec := resolve.ImportSpec{Lang: languageName, Imp: className.FullyQualifiedClassName()}
				if ol, found := resolve.FindRuleWithOverride(c, classImportSpec, languageName); found {
					labels.Add(simplifyLabel(c.RepoName, ol, from))
					resolvedAny = true
					continue
				}

				l, err := jr.lang.mavenResolver.ResolveClass(className, pc.ExcludedArtifacts(), pc.MavenRepositoryName())
				if err != nil {
					jr.lang.logger.Warn().Err(err).Str("class", className.FullyQualifiedClassName()).Msg("error resolving class")
					continue
				}
				if l == label.NoLabel {
					l = jr.resolveSingleClass(c, pc, className, ix, from, isTestRule)
				}
				if l != label.NoLabel {
					labels.Add(simplifyLabel(c.RepoName, l, from))
					resolvedAny = true
				}
			}

			if !resolvedAny {
				jr.lang.logger.Error().
					Str("package", imp.Name).
					Strs("classes", pkgClasses).
					Stringer("from", from).
					Msg("package has multiple providers and class-level resolution failed for all classes")
				jr.lang.hasHadErrors = true
			}
		}
	}

	setLabelAttrIncludingExistingValues(r, attrName, labels)

}

func (jr *Resolver) populatePluginsAttr(c *config.Config, ix *resolve.RuleIndex, resolveInput types.ResolveInput, packageConfig *javaconfig.Config, from label.Label, isTestRule bool, r *rule.Rule) {
	pluginLabels := sorted_set.NewSortedSetFn[label.Label]([]label.Label{}, labelLess)
	for _, annotationProcessor := range resolveInput.AnnotationProcessors.SortedSlice() {
		dep := jr.resolveSinglePackage(c, packageConfig, annotationProcessor.PackageName(), ix, from, isTestRule, resolveInput.PackageNames, []string{annotationProcessor.BareOuterClassName()})
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

// resolveSinglePackageWithAmbiguity resolves a package import and returns whether there was ambiguity.
// When ambiguous is true and out is NoLabel, the caller should attempt class-level resolution.
func (jr *Resolver) resolveSinglePackageWithAmbiguity(c *config.Config, pc *javaconfig.Config, imp types.PackageName, ix *resolve.RuleIndex, from label.Label, isTestRule bool, ownPackageNames *sorted_set.SortedSet[types.PackageName], pkgClasses []string) (out label.Label, ambiguous bool) {
	cacheKey := types.NewResolvableJavaPackage(imp, false, false)
	importSpec := resolve.ImportSpec{Lang: languageName, Imp: cacheKey.String()}
	if ol, found := resolve.FindRuleWithOverride(c, importSpec, languageName); found {
		return ol, false
	}

	matches := ix.FindRulesByImportWithConfig(c, importSpec, languageName)

	if pc.ResolveToJavaExports() {
		matches = jr.tryResolvingToJavaExport(matches, from)
	} else {
		nonExportMatches := make([]resolve.FindResult, 0)
		for _, match := range matches {
			if !jr.lang.javaExportIndex.IsJavaExport(match.Label) {
				nonExportMatches = append(nonExportMatches, match)
			}
		}
		matches = nonExportMatches
	}

	if len(matches) == 1 {
		return matches[0].Label, false
	}

	if len(matches) > 1 {
		// Multiple matches found - signal ambiguity so caller can try class-level resolution
		return label.NoLabel, true
	}

	if v, ok := jr.internalCache.Get(cacheKey); ok {
		return simplifyLabel(c.RepoName, v.(label.Label), from), false
	}

	jr.lang.logger.Debug().Str("parsedImport", imp.Name).Stringer("from", from).Msg("not found yet")

	defer func() {
		if out != label.NoLabel {
			jr.internalCache.Add(cacheKey, out)
		}
	}()

	if java.IsStdlib(imp) {
		return label.NoLabel, false
	}
	if kotlin.IsStdlib(imp) {
		return label.NoLabel, false
	}

	// As per https://github.com/bazelbuild/bazel/blob/347407a88fd480fc5e0fbd42cc8196e4356a690b/tools/java/runfiles/Runfiles.java#L41
	if imp.Name == "com.google.devtools.build.runfiles" {
		runfilesLabel := "@bazel_tools//tools/java/runfiles"
		l, err := label.Parse(runfilesLabel)
		if err != nil {
			jr.lang.logger.Fatal().Str("label", runfilesLabel).Err(err).Msg("failed to parse known-good runfiles label")
			return label.NoLabel, false
		}
		return l, false
	}

	if l, err := jr.lang.mavenResolver.Resolve(imp, pc.ExcludedArtifacts(), pc.MavenRepositoryName()); err != nil {
		var noExternal *maven.NoExternalImportsError
		var multipleExternal *maven.MultipleExternalImportsError

		if errors.As(err, &noExternal) {
			// do not fail, the package might be provided elsewhere
		} else if errors.As(err, &multipleExternal) {
			// Maven has multiple options (split package) - check if class-level resolution is available
			if len(pkgClasses) > 0 {
				// Only signal ambiguity if we have class index data for at least one class
				for _, className := range pkgClasses {
					cls := types.NewClassName(imp, className)
					if resolved, _ := jr.lang.mavenResolver.ResolveClass(cls, pc.ExcludedArtifacts(), pc.MavenRepositoryName()); resolved != label.NoLabel {
						return label.NoLabel, true
					}
				}
			}
			// No class-level resolution available, show helpful error with resolution hints
			jr.lang.logger.Error().Strs("classes", pkgClasses).Msg("Append one of the following to BUILD.bazel:")
			for _, possible := range multipleExternal.PossiblePackages {
				jr.lang.logger.Error().Msgf("# gazelle:resolve java %s %s", imp.Name, possible)
			}
			// Don't return here - let execution continue to produce the warning about unresolved package
		} else {
			jr.lang.logger.Fatal().Err(err).Msg("maven resolver error")
		}
	} else {
		return l, false
	}

	if isTestRule {
		// If there's exactly one testonly match, use it
		testonlyCacheKey := types.NewResolvableJavaPackage(imp, true, false)
		testonlyImportSpec := resolve.ImportSpec{Lang: languageName, Imp: testonlyCacheKey.String()}
		testonlyMatches := ix.FindRulesByImportWithConfig(c, testonlyImportSpec, languageName)
		if len(testonlyMatches) == 1 {
			cacheKey = testonlyCacheKey
			return simplifyLabel(c.RepoName, testonlyMatches[0].Label, from), false
		}

		// If there's exactly one testsuite match, use it
		testsuiteCacheKey := types.NewResolvableJavaPackage(imp, true, true)
		testsuiteImportSpec := resolve.ImportSpec{Lang: languageName, Imp: testsuiteCacheKey.String()}
		testsuiteMatches := ix.FindRulesByImportWithConfig(c, testsuiteImportSpec, languageName)
		if len(testsuiteMatches) == 1 {
			cacheKey = testsuiteCacheKey
			l := testsuiteMatches[0].Label
			if l != from {
				l.Name += "-test-lib"
				return simplifyLabel(c.RepoName, l, from), false
			}
		}
	}

	if isTestRule && ownPackageNames.Contains(imp) {
		// Tests may have unique packages which don't exist outside of those tests - don't treat this as an error.
		return label.NoLabel, false
	}

	jr.lang.logger.Error().
		Str("package", imp.Name).
		Str("from rule", from.String()).
		Strs("classes", pkgClasses).
		Msg("Unable to find package for import in any dependency")
	jr.lang.hasHadErrors = true

	return label.NoLabel, false
}

func (jr *Resolver) resolveSinglePackage(c *config.Config, pc *javaconfig.Config, imp types.PackageName, ix *resolve.RuleIndex, from label.Label, isTestRule bool, ownPackageNames *sorted_set.SortedSet[types.PackageName], pkgClasses []string) (out label.Label) {
	out, _ = jr.resolveSinglePackageWithAmbiguity(c, pc, imp, ix, from, isTestRule, ownPackageNames, pkgClasses)
	return out
}

// buildPackageClassIndex lazily builds a class index for a specific package.
// Only called when package-level resolution is ambiguous (split packages).
func (jr *Resolver) buildPackageClassIndex(c *config.Config, pkg types.PackageName, ix *resolve.RuleIndex) *packageClassIndex {
	if pci, ok := jr.classIndex[pkg]; ok {
		return pci
	}

	// Find all rules that provide this package
	cacheKey := types.NewResolvableJavaPackage(pkg, false, false)
	importSpec := resolve.ImportSpec{Lang: languageName, Imp: cacheKey.String()}
	matches := ix.FindRulesByImportWithConfig(c, importSpec, languageName)

	// Also check for testonly providers
	testCacheKey := types.NewResolvableJavaPackage(pkg, true, false)
	testImportSpec := resolve.ImportSpec{Lang: languageName, Imp: testCacheKey.String()}
	testMatches := ix.FindRulesByImportWithConfig(c, testImportSpec, languageName)
	matches = append(matches, testMatches...)

	pci := &packageClassIndex{
		prod: make(map[string][]label.Label),
		test: make(map[string][]label.Label),
	}

	for _, m := range matches {
		// Try lookup without repo prefix since that's how we store entries
		cacheLabel := label.New("", m.Label.Pkg, m.Label.Name)
		info, ok := jr.lang.classExportCache[cacheLabel.String()]
		if !ok {
			continue
		}
		for _, cls := range info.classes {
			if cls.PackageName() != pkg {
				continue
			}
			name := cls.BareOuterClassName()
			if info.testonly {
				pci.test[name] = append(pci.test[name], m.Label)
			} else {
				pci.prod[name] = append(pci.prod[name], m.Label)
			}
		}
	}

	jr.classIndex[pkg] = pci
	jr.lang.logger.Debug().
		Str("package", pkg.Name).
		Int("prod_classes", len(pci.prod)).
		Int("test_classes", len(pci.test)).
		Msg("built class index for split package")

	return pci
}

func (jr *Resolver) resolveSingleClass(c *config.Config, pc *javaconfig.Config, className types.ClassName, ix *resolve.RuleIndex, from label.Label, isTestRule bool) (out label.Label) {
	imp := className.FullyQualifiedClassName()
	// Check for manual override first
	importSpec := resolve.ImportSpec{Lang: languageName, Imp: imp}
	if ol, found := resolve.FindRuleWithOverride(c, importSpec, languageName); found {
		return ol
	}

	// Build/get the per-package class index
	pkg := className.PackageName()
	pci := jr.buildPackageClassIndex(c, pkg, ix)
	bareClassName := className.BareOuterClassName()

	// Look up candidates - prefer prod classes, but test rules can also use test classes
	var candidates []label.Label
	if prodCandidates, ok := pci.prod[bareClassName]; ok {
		candidates = prodCandidates
	}
	if isTestRule {
		if testCandidates, ok := pci.test[bareClassName]; ok {
			candidates = append(candidates, testCandidates...)
		}
	}

	if len(candidates) == 0 {
		// No in-repo provider for this class. Mirror Gazelle's index-then-CrossResolve
		// ordering at class granularity: consult external plugins via the cross-resolver.
		return jr.resolveClassFromCrossResolver(c, pc, className, ix, from)
	}

	if len(candidates) == 1 {
		return simplifyLabel(c.RepoName, candidates[0], from)
	}

	// Multiple candidates - try java_export narrowing
	if pc.ResolveToJavaExports() {
		results := make([]resolve.FindResult, 0, len(candidates))
		for _, l := range candidates {
			results = append(results, resolve.FindResult{Label: l})
		}
		narrowed := jr.tryResolvingToJavaExport(results, from)
		if len(narrowed) == 1 {
			return simplifyLabel(c.RepoName, narrowed[0].Label, from)
		}
	}

	// Still ambiguous - log error
	labels := make([]string, 0, len(candidates))
	for _, l := range candidates {
		labels = append(labels, l.String())
	}
	sort.Strings(labels)

	jr.lang.logger.Error().
		Str("class", imp).
		Strs("targets", labels).
		Msg("resolveSingleClass found MULTIPLE providers for class")

	return label.NoLabel
}

// ruleDeclaresClass reports whether the rule at lbl is known (via the class export
// cache) to provide the outer class of className. It lets the fast path skip
// cross-resolver probes for classes the resolved in-repo target already owns.
func (jr *Resolver) ruleDeclaresClass(lbl label.Label, className types.ClassName) bool {
	cacheLabel := label.New("", lbl.Pkg, lbl.Name)
	info, ok := jr.lang.classExportCache[cacheLabel.String()]
	if !ok {
		return false
	}
	for _, cls := range info.classes {
		if cls.PackageName() == className.PackageName() && cls.BareOuterClassName() == className.BareOuterClassName() {
			return true
		}
	}
	return false
}

// resolveClassFromCrossResolver consults registered Gazelle CrossResolvers for a
// class-level java import. Because the Java plugin intentionally never registers
// classes in the global RuleIndex, a class-level FindRulesByImportWithConfig always
// misses the index and falls through to CrossResolve. This lets external plugins
// (e.g. proto/wire generators) provide class-level resolutions even when an in-repo
// target owns the enclosing package (a split package).
// resolveTestSuiteHelperClass returns the "<suite>-test-lib" helper library of the lone
// java_test_suite that provides imp AND declares className, or NoLabel otherwise.
// java_test_suite moves its non-test sources into a helper library named <suite>-test-lib;
// a test in another package that imports one of those helpers must depend on it. This mirrors
// resolveSinglePackageWithAmbiguity's test-suite branch, which only fires when imp has no
// production provider -- here we cover the case where it has one (a split main/test-suite
// package), so the package itself resolves to production and only the helper class is missing.
// The className check guards against false positives: a class the production provider does not
// declare may be a main top-level function (not registered as a class), which the helper lib
// does not declare either and must not pull in.
func (jr *Resolver) resolveTestSuiteHelperClass(c *config.Config, imp types.PackageName, className types.ClassName, ix *resolve.RuleIndex, from label.Label) label.Label {
	spec := resolve.ImportSpec{Lang: languageName, Imp: types.NewResolvableJavaPackage(imp, true, true).String()}
	matches := ix.FindRulesByImportWithConfig(c, spec, languageName)
	if len(matches) != 1 {
		return label.NoLabel
	}
	l := matches[0].Label
	if l == from {
		return label.NoLabel
	}
	l.Name += "-test-lib"
	if !jr.ruleDeclaresClass(l, className) {
		return label.NoLabel
	}
	return simplifyLabel(c.RepoName, l, from)
}

func (jr *Resolver) resolveClassFromCrossResolver(c *config.Config, pc *javaconfig.Config, className types.ClassName, ix *resolve.RuleIndex, from label.Label) label.Label {
	importSpec := resolve.ImportSpec{Lang: languageName, Imp: className.FullyQualifiedClassName()}
	matches := ix.FindRulesByImportWithConfig(c, importSpec, languageName)
	if len(matches) == 0 {
		return label.NoLabel
	}

	if pc.ResolveToJavaExports() {
		matches = jr.tryResolvingToJavaExport(matches, from)
	}

	candidates := sorted_set.NewSortedSetFn[label.Label]([]label.Label{}, sorted_set.LabelLess)
	for _, m := range matches {
		candidates.Add(m.Label)
	}

	if candidates.Len() == 1 {
		return simplifyLabel(c.RepoName, candidates.SortedSlice()[0], from)
	}

	labelStrings := make([]string, 0, candidates.Len())
	for _, l := range candidates.SortedSlice() {
		labelStrings = append(labelStrings, l.String())
	}
	jr.lang.logger.Error().
		Str("class", className.FullyQualifiedClassName()).
		Strs("targets", labelStrings).
		Msg("cross-resolver returned multiple providers for class")
	return label.NoLabel
}

// tryResolvingToJavaExport attempts to narrow down a list of resolution candidates by preferring java_export targets when appropriate.
// A dependency will be resolved to a `java_export` target when the following are all true.
//   - The dependency is contained in a java_export target, and
//   - There is exactly one java_export target that contains the dependency, and
//   - That java_export does not export the target under consideration (`from`).
//
// Returns a subset of `results`, either by picking an appropriate `java_export`, or by eliminating ineligible `java_export`s.
// The program will issue a fatal error if it finds that more than one java_export contains the required dependency.
func (jr *Resolver) tryResolvingToJavaExport(results []resolve.FindResult, from label.Label) []resolve.FindResult {
	coveredByTheSameExport := func(one, other label.Label) bool {
		oneExport, oneIsCoveredByExport := jr.lang.javaExportIndex.IsExportedByJavaExport(one)
		otherExport, otherIsCoveredByExport := jr.lang.javaExportIndex.IsExportedByJavaExport(other)

		if !oneIsCoveredByExport && !otherIsCoveredByExport {
			return true
		} else if oneIsCoveredByExport && otherIsCoveredByExport {
			return oneExport.Label == otherExport.Label
		}
		return false
	}

	var javaExportsThatCoverThisDep []resolve.FindResult
	var nonJavaExportResults []resolve.FindResult
	for _, result := range results {
		if jr.lang.javaExportIndex.IsJavaExport(result.Label) {
			javaExportsThatCoverThisDep = append(javaExportsThatCoverThisDep, result)
		} else {
			if !coveredByTheSameExport(from, result.Label) {
				dependencyExporter, dependencyIsCovered := jr.lang.javaExportIndex.IsExportedByJavaExport(result.Label)
				if dependencyIsCovered {
					javaExportsThatCoverThisDep = append(javaExportsThatCoverThisDep, resolve.FindResult{Label: dependencyExporter.Label})
				}
			}
			nonJavaExportResults = append(nonJavaExportResults, result)
		}
	}

	if len(javaExportsThatCoverThisDep) == 0 {
		return results
	} else if len(javaExportsThatCoverThisDep) == 1 {
		return javaExportsThatCoverThisDep
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

	// If we don't find any relevant java_export, resolve normally.
	return nonJavaExportResults
}

func isJvmLibrary(c *config.Config, kind string) bool {
	return isJavaLibrary(c, kind) || isKotlinLibrary(kind)
}

func isJavaLibrary(c *config.Config, kind string) bool {
	return kind == "java_library" || isJavaProtoLibrary(c, kind)
}

func isKotlinLibrary(kind string) bool {
	return kind == "kt_jvm_library"
}

func isJavaProtoLibrary(c *config.Config, kind string) bool {
	javaProtoLibrary := "java_proto_library"
	javaGrpcLibrary := "java_grpc_library"

	// Check if this kind is mapped FROM a proto library via map_kind
	for _, mappedKind := range c.KindMap {
		if mappedKind.KindName == kind {
			if mappedKind.FromKind == javaProtoLibrary {
				javaProtoLibrary = kind
				break
			}

			if mappedKind.FromKind == javaGrpcLibrary {
				javaGrpcLibrary = kind
				break
			}
		}
	}

	return kind == javaProtoLibrary || kind == javaGrpcLibrary
}
