package java_export_index

import (
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/rs/zerolog"
	"sort"
)

// JavaExportResolveInfo captures metadata about a java_export rule.
// We could capture this in private attributes, but we need to access it from every java_library
// exported by the given java_export, and we can't access other rule.Rule instances during resolve.
type JavaExportResolveInfo struct {
	Rule               *rule.Rule
	Label              label.Label
	InternalVisibility *sorted_set.SortedSet[label.Label]
}

func NewJavaExportResolveInfoFromRule(repoName string, r *rule.Rule, file *rule.File) *JavaExportResolveInfo {
	lbl := label.New(repoName, file.Pkg, r.Name())
	exportPackageVisibility := label.New("", file.Pkg, "__pkg__")
	return &JavaExportResolveInfo{
		Rule:               r,
		Label:              lbl,
		InternalVisibility: sorted_set.NewSortedSetFn([]label.Label{exportPackageVisibility}, sorted_set.LabelLess),
	}
}

// JavaExportIndex holds information about `java_export` targets and which symbols they make available,
// so that other java targets can depend on the right `java_export` instead of fine-grained dependencies.
type JavaExportIndex struct {
	// langName and logger are fields we must store from the language.Language, so that we can use them in the implementation.
	langName string
	logger   zerolog.Logger

	// readyForResolve is an internal flag that will turn true when the index is ready to perform resolution.
	// It should be set after we've generated all the java_library targets, but before starting resolution.
	// Before this flag, it's expected that `javaExports` only contains sparse information
	readyForResolve bool

	// packagesToLabelsDeclaringThem and labelsToResolveInputs are used to calculate the transitive closure of `java_exports` targets.
	// They are filled out _during_ the `GenerateRules` phase, and used at the end to populate javaExports and labelToJavaExport.
	packagesToLabelsDeclaringThem map[types.PackageName]label.Label
	labelsToResolveInputs         map[label.Label]types.ResolveInput

	// javaExports and labelToJavaExport are used to resolve the dependencies of `java_library` targets,
	// to decide whether they're going to depend on a `java_export` or a fine-grained dependency.
	// They are filled _after_ the `GenerateRules` phase, and used during the `Resolve` phase.
	javaExports       map[label.Label]*JavaExportResolveInfo
	labelToJavaExport map[label.Label]label.Label
}

func NewJavaExportIndex(langName string, logger zerolog.Logger) *JavaExportIndex {
	return &JavaExportIndex{
		langName:                      langName,
		logger:                        logger,
		readyForResolve:               false,
		packagesToLabelsDeclaringThem: make(map[types.PackageName]label.Label),
		labelsToResolveInputs:         make(map[label.Label]types.ResolveInput),
		javaExports:                   make(map[label.Label]*JavaExportResolveInfo),
		labelToJavaExport:             make(map[label.Label]label.Label),
	}
}

// RecordRuleWithResolveInput lets the index know about a rule that might declare some packages, and might depend on some other packages later.
// Must be called before FinalizeIndex.
func (jei *JavaExportIndex) RecordRuleWithResolveInput(repoName string, file *rule.File, r *rule.Rule, resolveInput types.ResolveInput) {
	pkg := ""
	if file != nil {
		pkg = file.Pkg
	}
	lbl := label.New(repoName, pkg, r.Name())
	if jei.readyForResolve {
		jei.logger.Fatal().
			Str("label", lbl.String()).
			Msg("Tried to record rule after the index was finalized. This is likely an internal bug, please contact the maintainers.")
	}

	jei.labelsToResolveInputs[lbl] = resolveInput
	for _, javaPackage := range resolveInput.PackageNames.SortedSlice() {
		jei.packagesToLabelsDeclaringThem[javaPackage] = lbl
	}
}

// RecordJavaExport lets the index know about a java_export rule, for later resolution.
// Must be called before FinalizeIndex.
func (jei *JavaExportIndex) RecordJavaExport(repoName string, r *rule.Rule, f *rule.File) {
	lbl := label.New(repoName, f.Pkg, r.Name())
	if jei.readyForResolve {
		jei.logger.Fatal().
			Str("label", lbl.String()).
			Msg("Tried to record java_export after the index was finalized. This is likely an internal bug, please contact the maintainers.")
	}
	srcs := r.AttrStrings("srcs")
	if len(srcs) > 0 {
		jei.logger.Error().
			Str("label", lbl.String()).
			Msg("java_export rule contained a non-empty `srcs` attribute, but it will be ignored during resolution. Instead, please use the `exports` or `runtime_deps` attributes and depend on the generated `java_library`")
	}
	jei.javaExports[lbl] = NewJavaExportResolveInfoFromRule(repoName, r, f)
}

// FinalizeIndex processes all the `java_exports` we've recorded when traversing the repository, to:
// - Gather all the transitive dependencies by traversing the `ResolveInput`s of relevant targets.
// - With that information, populate the map of `labelToJavaExport`.
type exportConflict struct {
	dep                 label.Label
	wantedToExportFrom  label.Label
	alreadyExportedFrom label.Label
}

func (jei *JavaExportIndex) FinalizeIndex() {

	exportConflicts := sorted_set.NewSortedSetFn[exportConflict]([]exportConflict{}, func(a, b exportConflict) bool {
		return sorted_set.LabelLess(a.wantedToExportFrom, b.wantedToExportFrom)
	})

	for _, javaExport := range jei.javaExports {
		jei.calculateImportsForJavaExport(javaExport, exportConflicts)
	}

	// We need to collect and sort the conflicts to get a deterministic ordering of output for tests.
	for _, conflict := range exportConflicts.SortedSlice() {
		conflictingExports := []string{conflict.wantedToExportFrom.String(), conflict.alreadyExportedFrom.String()}
		sort.Strings(conflictingExports)

		jei.logger.Error().
			Str("dependency", conflict.dep.String()).
			Strs("java_exports", conflictingExports).
			Msg("Two `java_export` targets want to export the same dependency. This can lead to incorrect results, please disambiguate, e.g. by having export depend on other export explicitly.")
	}

	jei.readyForResolve = true
}

func (jei *JavaExportIndex) calculateImportsForJavaExport(javaExport *JavaExportResolveInfo, conflicts *sorted_set.SortedSet[exportConflict]) {
	var parseErrors []error
	deps, errors := attrLabels("deps", javaExport.Rule, javaExport.Label)
	parseErrors = append(parseErrors, errors...)
	exports, errors := attrLabels("exports", javaExport.Rule, javaExport.Label)
	parseErrors = append(parseErrors, errors...)
	runtimeDeps, errors := attrLabels("runtime_deps", javaExport.Rule, javaExport.Label)
	parseErrors = append(parseErrors, errors...)

	if len(parseErrors) > 0 {
		jei.logger.Error().
			Errs("errors", errors).
			Msgf("Errors parsing labels from fields of %s", javaExport.Label.String())
	}

	labelsToVisit := make([]label.Label, len(deps))
	_ = copy(labelsToVisit, deps)
	labelsToVisit = append(labelsToVisit, exports...)
	labelsToVisit = append(labelsToVisit, runtimeDeps...)

	for _, lbl := range labelsToVisit {
		if jei.IsJavaExport(lbl) {
			continue
		}
		exportingExport, isExportedByAnotherJavaExport := jei.isExportedByJavaExport(lbl)
		if isExportedByAnotherJavaExport {
			conflicts.Add(exportConflict{
				dep:                 lbl,
				wantedToExportFrom:  javaExport.Label,
				alreadyExportedFrom: exportingExport.Label,
			})
		}
	}

	transitiveDeps := make(map[label.Label]bool)
	for _, depLabel := range labelsToVisit {
		transitiveDeps[depLabel] = true
	}

	// Breadth-first traversal on the transitive closure of the export,
	// resolving all the packages to the labels that export them.
	for len(labelsToVisit) > 0 {
		dep := labelsToVisit[0]
		labelsToVisit = labelsToVisit[1:]

		if jei.IsJavaExport(dep) {
			continue
		}

		// Visit the dependency
		jei.labelToJavaExport[dep] = javaExport.Label
		visibilityLbl := label.New("", dep.Pkg, "__pkg__")
		javaExport.InternalVisibility.Add(visibilityLbl)

		// Queue every transitive dependency to be visited
		resolveInputForDep := jei.labelsToResolveInputs[dep]
		for _, importedPkg := range resolveInputForDep.ImportedPackageNames.SortedSlice() {
			lblToVisit, found := jei.packagesToLabelsDeclaringThem[importedPkg]
			if !found || lblToVisit == label.NoLabel {
				jei.logger.Debug().
					Str("package", importedPkg.Name).
					Msg("Found no label for imported java package. It's probably a standard library package, or a package from maven")
				continue
			}
			if jei.IsJavaExport(lblToVisit) {
				continue
			}
			exportingExport, isExportedByAnotherJavaExport := jei.isExportedByJavaExport(lblToVisit)
			if isExportedByAnotherJavaExport {
				conflicts.Add(exportConflict{
					dep:                 lblToVisit,
					wantedToExportFrom:  javaExport.Label,
					alreadyExportedFrom: exportingExport.Label,
				})
				continue
			}
			if ok := transitiveDeps[lblToVisit]; !ok {
				labelsToVisit = append(labelsToVisit, lblToVisit)
				transitiveDeps[lblToVisit] = true
			}
		}
	}
}

func (jei *JavaExportIndex) IsJavaExport(lbl label.Label) bool {
	_, is := jei.javaExports[lbl]
	return is
}

func (jei *JavaExportIndex) IsExportedByJavaExport(lbl label.Label) (*JavaExportResolveInfo, bool) {
	if !jei.readyForResolve {
		jei.logger.Fatal().
			Str("label", lbl.String()).
			Msg("Tried to get check if label is exported by a java_export before the java export index was ready for resolve. This is likely an internal bug, please contact the maintainers.")
	}
	return jei.isExportedByJavaExport(lbl)
}

// isExportedByJavaExport is the private version of IsExportedByJavaExport.
// It exists so that we can call it while we finish the index, while it's still not ready for resolution.
func (jei *JavaExportIndex) isExportedByJavaExport(lbl label.Label) (*JavaExportResolveInfo, bool) {
	exportLbl, isExported := jei.labelToJavaExport[lbl]
	if isExported {
		export, exists := jei.javaExports[exportLbl]
		if !exists {
			jei.logger.Fatal().
				Str("label", lbl.String()).
				Str("java_export", exportLbl.String()).
				Msg("Label is exported by java_export, but target is not recorded")
		}
		return export, true
	}
	return nil, false
}

// VisibilityForLabel returns the visibility that a target label.Label should have, according to the JavaExportIndex.
// Returns nil if the JavaExportIndex doesn't have an opinion on what visibility a target should have.
func (jei *JavaExportIndex) VisibilityForLabel(lbl label.Label) *sorted_set.SortedSet[label.Label] {
	if !jei.readyForResolve {
		jei.logger.Fatal().
			Str("target", lbl.String()).
			Msg("Tried to get visibility for target before the java export index was ready for resolve. This is likely an internal bug, please contact the maintainers.")
	}
	regularReturn := sorted_set.NewSortedSetFn[label.Label]([]label.Label{label.New("", "", "__subpackages__")}, sorted_set.LabelLess)
	if jei.IsJavaExport(lbl) {
		return regularReturn
	}

	exporter, isExportedByJavaExport := jei.IsExportedByJavaExport(lbl)
	if isExportedByJavaExport {
		return exporter.InternalVisibility
	}

	return nil
}

func attrLabels(attr string, r *rule.Rule, ruleLabel label.Label) ([]label.Label, []error) {
	depsStrings := r.AttrStrings(attr)
	deps := make([]label.Label, 0, len(depsStrings))
	errors := make([]error, 0)
	for _, depString := range depsStrings {
		lbl, err := label.Parse(depString)
		if err != nil {
			errors = append(errors, err)
		}
		if lbl.Pkg == "" {
			lbl.Pkg = ruleLabel.Pkg
			lbl.Relative = false
		}
		lbl.Repo = ruleLabel.Repo
		deps = append(deps, lbl)
	}
	return deps, errors
}
