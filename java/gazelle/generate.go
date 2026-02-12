package gazelle

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/rule"
	bzl "github.com/bazelbuild/buildtools/build"
	"github.com/rs/zerolog"
)

type javaFile struct {
	pathRelativeToBazelWorkspaceRoot string
	pkg                              types.PackageName
}

func (jf *javaFile) ClassName() *types.ClassName {
	name := filepath.Base(jf.pathRelativeToBazelWorkspaceRoot)
	if strings.HasSuffix(name, ".java") {
		name = strings.TrimSuffix(name, ".java")
	} else if strings.HasSuffix(name, ".kt") {
		name = strings.TrimSuffix(name, ".kt")
	}
	className := types.NewClassName(jf.pkg, name)
	return &className
}

func javaFileLess(l, r javaFile) bool {
	return l.pathRelativeToBazelWorkspaceRoot < r.pathRelativeToBazelWorkspaceRoot
}

type separateJavaTestReasons struct {
	attributes map[string]bzl.Expr
	wrapper    string
}

// GenerateRules extracts build metadata from source files in a directory.
//
// See language.GenerateRules for more information.
func (l javaLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	log := l.logger.With().Str("step", "GenerateRules").Str("rel", args.Rel).Logger()

	cfgs := args.Config.Exts[languageName].(javaconfig.Configs)
	cfg := cfgs[args.Rel]

	var res language.GenerateResult
	if !cfg.ExtensionEnabled() {
		return res
	}

	if cfg.GenerateProto() {
		generateProtoLibraries(&l, args, log, &res)
	}

	var srcFilenamesRelativeToPackage []string
	hasKotlinFiles := false
	if cfg.KotlinEnabled() {
		srcFilenamesRelativeToPackage = filterStrSlice(args.RegularFiles, func(f string) bool {
			ext := filepath.Ext(f)
			if ext == ".kt" {
				hasKotlinFiles = true
				return true
			} else {
				return ext == ".java"
			}
		})
	} else {
		srcFilenamesRelativeToPackage = filterStrSlice(args.RegularFiles, func(f string) bool { return filepath.Ext(f) == ".java" })
	}

	isResourcesRoot := strings.HasSuffix(args.Rel, "/resources")
	isResourcesSubdir := strings.Contains(args.Rel, "/resources/") && !isResourcesRoot
	isModule := cfg.ModuleGranularity() == "module"

	generateResources := cfg.GenerateResources()

	if !generateResources {
		// java_generate_resources == false: Disable resources logic
		isResourcesRoot = false
		isResourcesSubdir = false
	}

	var javaPkg *java.Package

	if len(srcFilenamesRelativeToPackage) == 0 {
		if isResourcesSubdir {
			// Skip subdirectories of resources roots - they shouldn't generate BUILD files
			return res
		}
		if !isResourcesRoot {
			if !isModule || !cfg.IsModuleRoot() {
				return res
			}
		}
		// For resources root directories, continue processing even without Java files
	}

	if len(srcFilenamesRelativeToPackage) == 0 && isResourcesRoot {
		// Skip Java parsing for resources-only directories
		javaPkg = &java.Package{
			Name: types.NewPackageName(""),
		}
	} else {
		sort.Strings(srcFilenamesRelativeToPackage)

		var err error
		javaPkg, err = l.parser.ParsePackage(context.Background(), &javaparser.ParsePackageRequest{
			Rel:   args.Rel,
			Files: srcFilenamesRelativeToPackage,
		})
		if err != nil {
			log.Fatal().Err(err).Str("package", args.Rel).Msg("Failed to parse package")
		}
	}

	// We exclude intra-package imports to avoid self-dependencies.
	// This isn't a great heuristic for a few reasons:
	//  1. We may want to split targets with more granularity than per-package.
	//  2. "What input files did you have" isn't a great heuristic for "What classes are generated"
	//     (e.g. inner classes, annotation processor generated classes, etc).
	// But it will do for now.
	likelyLocalClassNames := sorted_set.NewSortedSet([]string{})
	for _, filename := range srcFilenamesRelativeToPackage {
		if strings.HasSuffix(filename, ".kt") {
			fileWithoutExtension := strings.TrimSuffix(filename, ".kt")
			likelyLocalClassNames.Add(fileWithoutExtension)
			// Top level values and functions in Kotlin are accessible from Java under the <filename>Kt class.
			likelyLocalClassNames.Add(fileWithoutExtension + "Kt")
		} else {
			likelyLocalClassNames.Add(strings.TrimSuffix(filename, ".java"))
		}
	}

	if isModule {
		if len(srcFilenamesRelativeToPackage) > 0 {
			l.javaPackageCache[args.Rel] = javaPkg
		}

		if !cfg.IsModuleRoot() {
			log.Debug().Msg("module // sub directory, returning early")
			if args.File != nil {
				// In module mode, there should be no intermediate build files.
				if err := os.RemoveAll(args.File.Path); err != nil {
					log.Fatal().Err(err).Msg("could not delete build file")
				}
			}
			return res
		}
	} else if cfg.ParentModuleGranularity() == "module" {
		// This directory has explicitly set package granularity within a module.
		// Process it as a standalone package (don't aggregate into the parent module).
		log.Debug().Msg("package granularity override within module, processing as standalone package")
	}

	allMains := sorted_set.NewSortedSetFn[types.ClassName]([]types.ClassName{}, types.ClassNameLess)

	// Files and imports for code which isn't tests, and isn't used as helpers in tests.
	productionJavaFiles := sorted_set.NewSortedSet([]string{})
	productionJavaImports := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	productionJavaImportedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	nonLocalJavaExports := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	nonLocalJavaExportedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)

	// Files and imports for actual test classes.
	testJavaFiles := sorted_set.NewSortedSetFn([]javaFile{}, javaFileLess)
	testJavaImports := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	testJavaImportedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)

	// Java Test files which need to be generated separately from any others because they have explicit attribute overrides.
	separateTestJavaFiles := make(map[javaFile]separateJavaTestReasons)

	// Files which are used by non-test classes in test java packages.
	testHelperJavaFiles := sorted_set.NewSortedSetFn([]javaFile{}, javaFileLess)

	// All java packages present in this bazel package.
	allPackageNames := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)

	annotationProcessorClasses := sorted_set.NewSortedSetFn(nil, types.ClassNameLess)

	if isModule {
		for mRel, mJavaPkg := range l.javaPackageCache {
			if !strings.HasPrefix(mRel, args.Rel) {
				continue
			}
			allPackageNames.Add(mJavaPkg.Name)

			if !mJavaPkg.TestPackage {
				addNonLocalImportsAndExports(productionJavaImports, productionJavaImportedClasses, nonLocalJavaExports, nil, mJavaPkg.ImportedClasses, mJavaPkg.ImportedPackagesWithoutSpecificClasses, mJavaPkg.ExportedClasses, mJavaPkg.Name, likelyLocalClassNames)
				for _, f := range mJavaPkg.Files.SortedSlice() {
					productionJavaFiles.Add(filepath.Join(mRel, f))
					jf := javaFile{pathRelativeToBazelWorkspaceRoot: filepath.Join(mRel, f), pkg: mJavaPkg.Name}
					nonLocalJavaExportedClasses.Add(*jf.ClassName())
				}
				allMains.AddAll(mJavaPkg.Mains)
			} else {
				// Tests don't get to export things, as things shouldn't depend on them.
				addNonLocalImportsAndExports(testJavaImports, testJavaImportedClasses, nil, nil, mJavaPkg.ImportedClasses, mJavaPkg.ImportedPackagesWithoutSpecificClasses, mJavaPkg.ExportedClasses, mJavaPkg.Name, likelyLocalClassNames)
				for _, f := range mJavaPkg.Files.SortedSlice() {
					path := filepath.Join(mRel, f)
					file := javaFile{
						pathRelativeToBazelWorkspaceRoot: path,
						pkg:                              mJavaPkg.Name,
					}
					accumulateJavaFile(cfg, testJavaFiles, testHelperJavaFiles, separateTestJavaFiles, file, mJavaPkg.PerClassMetadata, log)
				}
			}
			for _, annotationClass := range mJavaPkg.AllAnnotations().SortedSlice() {
				annotationProcessorClasses.AddAll(cfg.GetAnnotationProcessorPluginClasses(annotationClass))
			}
		}
	} else {
		allPackageNames.Add(javaPkg.Name)
		if javaPkg.TestPackage {
			// Tests don't get to export things, as things shouldn't depend on them.
			addNonLocalImportsAndExports(testJavaImports, testJavaImportedClasses, nil, nil, javaPkg.ImportedClasses, javaPkg.ImportedPackagesWithoutSpecificClasses, javaPkg.ExportedClasses, javaPkg.Name, likelyLocalClassNames)
		} else {
			addNonLocalImportsAndExports(productionJavaImports, productionJavaImportedClasses, nonLocalJavaExports, nil, javaPkg.ImportedClasses, javaPkg.ImportedPackagesWithoutSpecificClasses, javaPkg.ExportedClasses, javaPkg.Name, likelyLocalClassNames)
		}
		allMains.AddAll(javaPkg.Mains)
		for _, f := range srcFilenamesRelativeToPackage {
			path := filepath.Join(args.Rel, f)
			if javaPkg.TestPackage {
				file := javaFile{
					pathRelativeToBazelWorkspaceRoot: path,
					pkg:                              javaPkg.Name,
				}
				accumulateJavaFile(cfg, testJavaFiles, testHelperJavaFiles, separateTestJavaFiles, file, javaPkg.PerClassMetadata, log)
			} else {
				productionJavaFiles.Add(path)
				jf := javaFile{pathRelativeToBazelWorkspaceRoot: path, pkg: javaPkg.Name}
				nonLocalJavaExportedClasses.Add(*jf.ClassName())
			}
		}
		for _, annotationClass := range javaPkg.AllAnnotations().SortedSlice() {
			annotationProcessorClasses.AddAll(cfg.GetAnnotationProcessorPluginClasses(annotationClass))
		}
	}

	allPackageNamesSlice := allPackageNames.SortedSlice()
	nonLocalProductionJavaImports := productionJavaImports.Filter(func(i types.PackageName) bool {
		for _, n := range allPackageNamesSlice {
			if i.Name == n.Name {
				return false
			}
		}
		return true
	})
	nonLocalProductionJavaImportedClasses := productionJavaImportedClasses.Filter(func(c types.ClassName) bool {
		for _, n := range allPackageNamesSlice {
			if c.PackageName().Name == n.Name {
				return false
			}
		}
		return true
	})

	javaLibraryKind := "java_library"
	if hasKotlinFiles {
		javaLibraryKind = "kt_jvm_library"
	}

	// Check if this is a resources root directory and generate a pkg_files target
	if isResourcesRoot && len(srcFilenamesRelativeToPackage) == 0 {
		// Collect resource files recursively from this directory and all subdirectories
		var allResourceFiles []string

		collectResourceFiles := func(files []string, dirPrefix string) []string {
			var resourceFiles []string
			for _, f := range files {
				base := filepath.Base(f)
				// Skip Java files, BUILD files, and common non-resource files
				if base == "BUILD" || base == "BUILD.bazel" { // files from our tests
					continue
				}
				if dirPrefix != "" {
					resourceFiles = append(resourceFiles, path.Join(dirPrefix, f))
				} else {
					resourceFiles = append(resourceFiles, f)
				}
			}
			return resourceFiles
		}

		allResourceFiles = append(allResourceFiles, collectResourceFiles(args.RegularFiles, "")...)

		for _, subdir := range args.Subdirs {
			// Skip BUILD directories
			if subdir == "BUILD" || subdir == "BUILD.bazel" {
				continue
			}
			subdirFiles := collectResourceFilesRecursively(args, subdir)
			allResourceFiles = append(allResourceFiles, subdirFiles...)
		}

		if len(allResourceFiles) > 0 {
			// Sort the files for deterministic output
			sort.Strings(allResourceFiles)

			// Always generate a pkg_files target for resources
			r := rule.NewRule("pkg_files", "resources")
			r.SetAttr("srcs", allResourceFiles)

			stripPrefix := cfg.StripResourcesPrefix()
			if stripPrefix != "" {
				r.SetAttr("strip_prefix", stripPrefix)
			}

			res.Gen = append(res.Gen, r)
			res.Imports = append(res.Imports, types.ResolveInput{})

			// In package mode, also generate a java_library wrapper for the resources
			if !isModule {
				resourceLib := rule.NewRule(javaLibraryKind, "resources_lib")
				resourceLib.SetAttr("resources", []string{":resources"})
				resourceLib.SetAttr("visibility", []string{"//:__subpackages__"})
				res.Gen = append(res.Gen, resourceLib)
				res.Imports = append(res.Imports, types.ResolveInput{})
			}
		}
	} else if productionJavaFiles.Len() > 0 {
		var resourcesDirectRef string  // For module mode: direct reference to pkg_files
		var resourcesRuntimeDep string // For package mode: runtime_dep on resources_lib

		if cfg.SourcesetRoot() != "" {
			// We have a sourceset root configured
			// The sourceset root is the parent of both java and resources directories
			// For example, if sourceset root is "src/sample", then:
			// - Java files are in "src/sample/java/..."
			// - Resources are in "src/sample/resources"

			if generateResources {
				resourcesPath := path.Join(cfg.SourcesetRoot(), "resources")

				// Check if the resources directory actually exists
				fullResourcesPath := filepath.Join(args.Config.RepoRoot, filepath.FromSlash(resourcesPath))
				if _, err := os.Stat(fullResourcesPath); err == nil {
					// Resources directory exists, add the reference
					if isModule {
						// Module mode: reference pkg_files directly as resources
						resourcesDirectRef = "//" + resourcesPath + ":resources"
					} else {
						// Package mode: reference resources_lib as runtime_deps
						resourcesRuntimeDep = "//" + resourcesPath + ":resources_lib"
					}
				}
			}
		}

		l.generateJavaLibrary(args.File, args.Rel, filepath.Base(args.Rel), productionJavaFiles.SortedSlice(), resourcesDirectRef, resourcesRuntimeDep, allPackageNames, nonLocalProductionJavaImports, nonLocalProductionJavaImportedClasses, nonLocalJavaExports, nonLocalJavaExportedClasses, annotationProcessorClasses, false, javaLibraryKind, &res, cfg, args.Config.RepoName)
	}

	if cfg.GenerateBinary() {
		l.processJavaBinary(args.File, args.Rel, allMains, testHelperJavaFiles, &res)
	}

	// We add special packages to point to testonly libraries which - this accumulates them,
	// as well as the existing java imports of tests.
	testJavaImportsWithHelpers := testJavaImports.Clone()
	testJavaImportedClassesWithHelpers := testJavaImportedClasses.Clone()

	if testHelperJavaFiles.Len() > 0 {
		// Suites generate their own helper library.
		if cfg.TestMode() == "file" {
			srcs := make([]string, 0, testHelperJavaFiles.Len())
			packages := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)

			for _, tf := range testHelperJavaFiles.SortedSlice() {
				packages.Add(tf.pkg)
				testJavaImportsWithHelpers.Add(tf.pkg)
				testJavaImportedClassesWithHelpers.Add(*tf.ClassName())
				srcs = append(srcs, tf.pathRelativeToBazelWorkspaceRoot)
			}
			// Test helper libraries typically don't have resources
			l.generateJavaLibrary(args.File, args.Rel, filepath.Base(args.Rel), srcs, "", "", packages, testJavaImports, testJavaImportedClasses, nonLocalJavaExports, nonLocalJavaExportedClasses, annotationProcessorClasses, true, javaLibraryKind, &res, cfg, args.Config.RepoName)
		}
	}

	allTestRelatedSrcs := testJavaFiles.Clone()
	allTestRelatedSrcs.AddAll(testHelperJavaFiles)

	if allTestRelatedSrcs.Len() > 0 {
		switch cfg.TestMode() {
		case "file":
			for _, tf := range testJavaFiles.SortedSlice() {
				separateJavaTestReasons := separateTestJavaFiles[tf]
				l.generateJavaTest(args.File, args.Rel, cfg.MavenRepositoryName(), tf, isModule, testJavaImportsWithHelpers, testJavaImportedClassesWithHelpers, annotationProcessorClasses, nil, separateJavaTestReasons.wrapper, separateJavaTestReasons.attributes, &res)
			}

		case "suite":
			packageNames := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
			for _, tf := range allTestRelatedSrcs.SortedSlice() {
				packageNames.Add(tf.pkg)
			}

			suiteName := filepath.Base(args.Rel)
			if isModule {
				suiteName += "-tests"
			}

			srcs := make([]string, 0, allTestRelatedSrcs.Len())
			for _, src := range allTestRelatedSrcs.SortedSlice() {
				if _, ok := separateTestJavaFiles[src]; !ok {
					srcs = append(srcs, strings.TrimPrefix(filepath.ToSlash(src.pathRelativeToBazelWorkspaceRoot), args.Rel+"/"))
				}
			}
			sort.Strings(srcs)
			if len(srcs) > 0 {
				l.generateJavaTestSuite(
					args.File,
					suiteName,
					srcs,
					packageNames,
					cfg.MavenRepositoryName(),
					testJavaImportsWithHelpers,
					testJavaImportedClassesWithHelpers,
					annotationProcessorClasses,
					cfg.GetCustomJavaTestFileSuffixes(),
					testHelperJavaFiles.Len() > 0,
					&res,
				)
			}

			sortedSeparateTestJavaFiles := sorted_set.NewSortedSetFn([]javaFile{}, javaFileLess)
			for src := range separateTestJavaFiles {
				sortedSeparateTestJavaFiles.Add(src)
			}
			for _, src := range sortedSeparateTestJavaFiles.SortedSlice() {
				var testHelperDep *string
				if testHelperJavaFiles.Len() > 0 {
					testHelperDep = ptr(testHelperLibname(suiteName))
				}
				separateJavaTestReasons := separateTestJavaFiles[src]
				l.generateJavaTest(args.File, args.Rel, cfg.MavenRepositoryName(), src, isModule, testJavaImportsWithHelpers, testJavaImportedClassesWithHelpers, annotationProcessorClasses, testHelperDep, separateJavaTestReasons.wrapper, separateJavaTestReasons.attributes, &res)
			}
		}
	}

	for i := 0; i < len(res.Gen); i++ {
		log.Debug().Fields(map[string]interface{}{
			"idx":     i,
			"rule":    fmt.Sprintf("%#v", res.Gen[i]),
			"imports": fmt.Sprintf("%#v", res.Imports[i]),
		}).Msg("generate return")
	}

	return res
}

func (l javaLang) collectRuntimeDeps(kind, name string, file *rule.File) *sorted_set.SortedSet[label.Label] {
	runtimeDeps := sorted_set.NewSortedSetFn([]label.Label{}, labelLess)
	if file == nil {
		return runtimeDeps
	}

	for _, r := range file.Rules {
		if r.Name() != name {
			continue
		}

		// This does not support non string list values from runtime_deps.
		// Currently, that means if a target has a runtime_deps of a different
		// kind (e.g. a select), we will remove it. Hopefully in the future we
		// can be less destructive.
		for _, dep := range r.AttrStrings("runtime_deps") {
			parsedLabel, err := label.Parse(dep)
			if err != nil {
				l.logger.Fatal().
					Str("file.Pkg", file.Pkg).
					Str("name", name).
					Str("dep", dep).
					Err(err).
					Msg("label parse error")
			}
			runtimeDeps.Add(parsedLabel)
		}
		break
	}

	return runtimeDeps
}

func generateProtoLibraries(l *javaLang, args language.GenerateArgs, log zerolog.Logger, res *language.GenerateResult) {
	var protoRuleNames []string
	protoPackages := make(map[string]proto.Package)
	protoFileInfo := make(map[string]proto.FileInfo)
	for _, r := range args.OtherGen {
		if r.Kind() != "proto_library" {
			continue
		}
		pkg := r.PrivateAttr(proto.PackageKey).(proto.Package)
		protoPackages[r.Name()] = pkg
		for name, info := range pkg.Files {
			protoFileInfo[name] = info
		}
		protoRuleNames = append(protoRuleNames, r.Name())
	}
	sort.Strings(protoRuleNames)

	for _, protoRuleName := range protoRuleNames {
		protoPackage := protoPackages[protoRuleName]

		jplName := strings.TrimSuffix(protoRuleName, "_proto") + "_java_proto"
		jglName := strings.TrimSuffix(protoRuleName, "_proto") + "_java_grpc"
		jlName := strings.TrimSuffix(protoRuleName, "_proto") + "_java_library"

		rjpl := rule.NewRule("java_proto_library", jplName)
		rjpl.SetAttr("deps", []string{":" + protoRuleName})
		res.Gen = append(res.Gen, rjpl)
		res.Imports = append(res.Imports, types.ResolveInput{})

		if protoPackage.HasServices {
			r := rule.NewRule("java_grpc_library", jglName)
			r.SetAttr("srcs", []string{":" + protoRuleName})
			r.SetAttr("deps", []string{":" + jplName})
			res.Gen = append(res.Gen, r)
			res.Imports = append(res.Imports, types.ResolveInput{})
		}

		rjl := rule.NewRule("java_library", jlName)
		rjl.SetAttr("visibility", []string{"//:__subpackages__"})
		var exports []string
		if protoPackage.HasServices {
			exports = append(exports, ":"+jglName)
		}
		rjl.SetAttr("exports", append(exports, ":"+jplName))
		packageName := types.NewPackageName(protoPackage.Options["java_package"])
		log.Debug().Str("pkg", packageName.Name).Msg("adding the proto import statement")
		rjl.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{*types.NewResolvableJavaPackage(packageName, false, false)})

		// Extract class names from proto files for class-level resolution.
		// Proto compilation generates Java classes for each message, enum, service,
		// and an outer class (named after the proto file or via java_outer_classname option).
		var protoClasses []types.ClassName
		for _, fileInfo := range protoPackage.Files {
			// Add the outer class name (container for all types in the proto file)
			outerClassName := protoOuterClassName(fileInfo)
			if outerClassName != "" {
				protoClasses = append(protoClasses, types.NewClassName(packageName, outerClassName))
			}
			for _, msg := range fileInfo.Messages {
				protoClasses = append(protoClasses, types.NewClassName(packageName, msg))
			}
			for _, enum := range fileInfo.Enums {
				protoClasses = append(protoClasses, types.NewClassName(packageName, enum))
			}
			for _, svc := range fileInfo.Services {
				protoClasses = append(protoClasses, types.NewClassName(packageName, svc))
			}
		}
		if len(protoClasses) > 0 {
			rjl.SetPrivateAttr(classesKey, protoClasses)
			ruleLabel := label.New("", args.Rel, jlName)
			l.classExportCache[ruleLabel.String()] = classExportInfo{
				classes:  protoClasses,
				testonly: false,
			}
			classNames := make([]string, 0, len(protoClasses))
			for _, c := range protoClasses {
				classNames = append(classNames, c.BareOuterClassName())
			}
			log.Debug().
				Str("rule", jlName).
				Str("label", ruleLabel.String()).
				Strs("classes", classNames).
				Msg("registered proto classes for class-level resolution")
		}

		res.Gen = append(res.Gen, rjl)
		res.Imports = append(res.Imports, types.ResolveInput{
			PackageNames: sorted_set.NewSortedSetFn([]types.PackageName{packageName}, types.PackageNameLess),
		})
	}
}

// protoOuterClassName returns the outer class name for a proto file.
// This is either explicitly set via java_outer_classname option, or derived from the file name.
func protoOuterClassName(fileInfo proto.FileInfo) string {
	// Check for explicit java_outer_classname option
	for _, opt := range fileInfo.Options {
		if opt.Key == "java_outer_classname" {
			return opt.Value
		}
	}
	// Default: derive from file name (e.g., "http.proto" -> "Http")
	name := fileInfo.Name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".proto")
	if name == "" {
		return ""
	}
	// Convert to PascalCase (capitalize first letter, handle underscores)
	return snakeToPascalCase(name)
}

// snakeToPascalCase converts a snake_case string to PascalCase.
func snakeToPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// We exclude intra-target imports because otherwise we'd get self-dependencies come resolve time.
// toExports is optional and may be nil. All other parameters are required and must be non-nil.
func addNonLocalImportsAndExports(toImports *sorted_set.SortedSet[types.PackageName], toImportedClasses *sorted_set.SortedSet[types.ClassName], toExports *sorted_set.SortedSet[types.PackageName], toExportedClasses *sorted_set.SortedSet[types.ClassName], fromImportedClasses *sorted_set.SortedSet[types.ClassName], fromPackages *sorted_set.SortedSet[types.PackageName], fromExportedClasses *sorted_set.SortedSet[types.ClassName], pkg types.PackageName, localClasses *sorted_set.SortedSet[string]) {
	toImports.AddAll(fromPackages)
	addFilteringOutOwnPackage(toImports, toImportedClasses, fromImportedClasses, pkg, localClasses)
	if toExports != nil {
		addFilteringOutOwnPackage(toExports, toExportedClasses, fromExportedClasses, pkg, localClasses)
	}
}

func addFilteringOutOwnPackage(to *sorted_set.SortedSet[types.PackageName], toClasses *sorted_set.SortedSet[types.ClassName], from *sorted_set.SortedSet[types.ClassName], ownPackage types.PackageName, localOuterClassNames *sorted_set.SortedSet[string]) {
	for _, fromPackage := range from.SortedSlice() {
		if ownPackage == fromPackage.PackageName() {
			if localOuterClassNames.Contains(fromPackage.BareOuterClassName()) {
				continue
			}
		}

		if fromPackage.PackageName().Name == "" {
			continue
		}

		to.Add(fromPackage.PackageName())
		if toClasses != nil {
			toClasses.Add(fromPackage)
		}
	}
}

func accumulateJavaFile(cfg *javaconfig.Config, testJavaFiles, testHelperJavaFiles *sorted_set.SortedSet[javaFile], separateTestJavaFiles map[javaFile]separateJavaTestReasons, file javaFile, perClassMetadata map[string]java.PerClassMetadata, log zerolog.Logger) {
	if cfg.IsJavaTestFile(filepath.Base(file.pathRelativeToBazelWorkspaceRoot)) {
		annotationClassNames := sorted_set.NewSortedSetFn[types.ClassName](nil, types.ClassNameLess)
		// We attribute annotations on inner classes as if they apply to the outer class, so we need to strip inner class names when comparing.
		for class, metadataForClass := range perClassMetadata {
			className, err := types.ParseClassName(class)
			if err != nil {
				log.Warn().Err(err).Str("class-name", class).Msg("Failed to parse class name which was seen to have an annotation")
				continue
			}
			if className.FullyQualifiedOuterClassName() == file.ClassName().FullyQualifiedOuterClassName() {
				annotationClassNames.AddAll(metadataForClass.AnnotationClassNames)
				for _, key := range metadataForClass.MethodAnnotationClassNames.Keys() {
					annotationClassNames.AddAll(metadataForClass.MethodAnnotationClassNames.Values(key))
				}
			}
		}

		perFileAttrs := make(map[string]bzl.Expr)
		wrapper := ""
		for _, annotationClassName := range annotationClassNames.SortedSlice() {
			if attrs, ok := cfg.AttributesForAnnotation(annotationClassName.FullyQualifiedClassName()); ok {
				for k, v := range attrs {
					if old, ok := perFileAttrs[k]; ok {
						log.Error().Str("file", file.pathRelativeToBazelWorkspaceRoot).Msgf("Saw conflicting attr overrides from annotations for attribute %v: %v and %v. Picking one at random.", k, old, v)
					}
					perFileAttrs[k] = v
				}
			}
			newWrapper, ok := cfg.WrapperForAnnotation(annotationClassName.FullyQualifiedClassName())
			if ok {
				if wrapper != "" {
					log.Error().Str("file", file.pathRelativeToBazelWorkspaceRoot).Msgf("Saw conflicting wrappers from annotations: %v and %v. Picking one at random.", wrapper, newWrapper)
				}
				wrapper = newWrapper
			}
		}
		testJavaFiles.Add(file)
		if len(perFileAttrs) > 0 || wrapper != "" {
			separateTestJavaFiles[file] = separateJavaTestReasons{
				attributes: perFileAttrs,
				wrapper:    wrapper,
			}
		}
	} else {
		testHelperJavaFiles.Add(file)
	}
}

func (l javaLang) generateJavaLibrary(file *rule.File, pathToPackageRelativeToBazelWorkspace, name string, srcsRelativeToBazelWorkspace []string, resourcesDirectRef string, resourcesRuntimeDep string, packages, imports *sorted_set.SortedSet[types.PackageName], importedClasses *sorted_set.SortedSet[types.ClassName], exports *sorted_set.SortedSet[types.PackageName], exportedClasses *sorted_set.SortedSet[types.ClassName], annotationProcessorClasses *sorted_set.SortedSet[types.ClassName], testonly bool, javaLibraryRuleKind string, res *language.GenerateResult, cfg *javaconfig.Config, repoName string) {
	r := rule.NewRule(javaLibraryRuleKind, name)

	srcs := make([]string, 0, len(srcsRelativeToBazelWorkspace))
	for _, src := range srcsRelativeToBazelWorkspace {
		srcs = append(srcs, strings.TrimPrefix(filepath.ToSlash(src), filepath.ToSlash(pathToPackageRelativeToBazelWorkspace+"/")))
	}
	sort.Strings(srcs)

	// Handle resources based on mode
	if resourcesDirectRef != "" {
		// Module mode: add resources directly to the library
		r.SetAttr("resources", []string{resourcesDirectRef})
	}

	// This is so we would default ALL runtime_deps to "keep" mode
	runtimeDeps := l.collectRuntimeDeps(javaLibraryRuleKind, name, file)

	// Package mode: add resources_lib as runtime_deps
	if resourcesRuntimeDep != "" {
		parsedLabel, err := label.Parse(resourcesRuntimeDep)
		if err != nil {
			l.logger.Error().Err(err).Str("label", resourcesRuntimeDep).Msg("Failed to parse resources runtime dep label")
		} else {
			runtimeDeps.Add(parsedLabel)
		}
	}

	if runtimeDeps.Len() > 0 {
		r.SetAttr("runtime_deps", labelsToStrings(runtimeDeps.SortedSlice()))
	}

	r.SetAttr("srcs", srcs)
	if testonly {
		r.SetAttr("testonly", true)
	} else {
		r.SetAttr("visibility", []string{"//:__subpackages__"})
	}

	resolvablePackages := make([]types.ResolvableJavaPackage, 0, packages.Len())
	for _, pkg := range packages.SortedSlice() {
		resolvablePackages = append(resolvablePackages, *types.NewResolvableJavaPackage(pkg, testonly, false))
	}
	r.SetPrivateAttr(packagesKey, resolvablePackages)
	if exportedClasses != nil {
		classes := exportedClasses.SortedSlice()
		r.SetPrivateAttr(classesKey, classes)
		// Cache the classes for class-level resolution during the resolve phase
		ruleLabel := label.New("", pathToPackageRelativeToBazelWorkspace, name)
		l.classExportCache[ruleLabel.String()] = classExportInfo{
			classes:  classes,
			testonly: testonly,
		}
	}
	res.Gen = append(res.Gen, r)

	resolveInput := types.ResolveInput{
		PackageNames:         packages,
		ImportedPackageNames: imports,
		ImportedClasses:      importedClasses,
		ExportedPackageNames: exports,
		ExportedClassNames:   exportedClasses,
		AnnotationProcessors: annotationProcessorClasses,
	}
	res.Imports = append(res.Imports, resolveInput)

	if cfg.ResolveToJavaExports() {
		l.javaExportIndex.RecordRuleWithResolveInput(repoName, file, r, resolveInput)
	}
}

func (l javaLang) processJavaBinary(file *rule.File, rel string, allMains *sorted_set.SortedSet[types.ClassName], testHelperJavaFiles *sorted_set.SortedSet[javaFile], res *language.GenerateResult) {
	var testHelperJavaClasses *sorted_set.SortedSet[types.ClassName]
	for _, m := range allMains.SortedSlice() {
		// Lazily populate because java_binaries are pretty rare
		if testHelperJavaClasses == nil {
			testHelperJavaClasses = sorted_set.NewSortedSetFn[types.ClassName]([]types.ClassName{}, types.ClassNameLess)
			for _, testHelperJavaFile := range testHelperJavaFiles.SortedSlice() {
				testHelperJavaClasses.Add(*testHelperJavaFile.ClassName())
			}
		}
		isTestOnly := false
		libName := filepath.Base(rel)
		if testHelperJavaClasses.Contains(m) {
			isTestOnly = true
			libName = testHelperLibname(libName)
		}
		l.generateJavaBinary(file, m, libName, isTestOnly, res)
	}
}

func (l javaLang) generateJavaBinary(file *rule.File, m types.ClassName, libName string, testonly bool, res *language.GenerateResult) {
	const ruleKind = "java_binary"
	name := m.BareOuterClassName()
	r := rule.NewRule("java_binary", name) // FIXME check collision on name
	r.SetAttr("main_class", m.FullyQualifiedClassName())

	if testonly {
		r.SetAttr("testonly", true)
	}

	runtimeDeps := l.collectRuntimeDeps(ruleKind, name, file)
	runtimeDeps.Add(label.Label{Name: libName, Relative: true})
	r.SetAttr("runtime_deps", labelsToStrings(runtimeDeps.SortedSlice()))
	r.SetAttr("visibility", []string{"//visibility:public"})
	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, types.ResolveInput{
		PackageNames: sorted_set.NewSortedSetFn([]types.PackageName{m.PackageName()}, types.PackageNameLess),
	})
}

func (l javaLang) generateJavaTest(file *rule.File, pathToPackageRelativeToBazelWorkspace string, mavenRepositoryName string, f javaFile, includePackageInName bool, imports *sorted_set.SortedSet[types.PackageName], importedClasses *sorted_set.SortedSet[types.ClassName], annotationProcessorClasses *sorted_set.SortedSet[types.ClassName], depOnTestHelpers *string, wrapper string, extraAttributes map[string]bzl.Expr, res *language.GenerateResult) {
	className := f.ClassName()
	fullyQualifiedTestClass := className.FullyQualifiedClassName()
	var testName string
	if includePackageInName {
		testName = strings.ReplaceAll(fullyQualifiedTestClass, ".", "_")
	} else {
		testName = className.BareOuterClassName()
	}

	javaRuleKind := "java_test"
	if importsJunit5(imports) {
		javaRuleKind = "java_junit5_test"
	}

	runtimeDeps := l.collectRuntimeDeps(javaRuleKind, testName, file)
	if importsJunit5(imports) {
		// This should probably register imports here, and then allow the
		// resolver to resolve this to an artifact, but we don't currently wire
		// up the resolver to do this. We probably should.
		// In the mean time, hard-code some labels.
		for _, artifact := range junit5RuntimeDeps {
			runtimeDeps.Add(maven.LabelFromArtifact(mavenRepositoryName, artifact))
		}
	}

	ruleKind := javaRuleKind
	if wrapper != "" {
		ruleKind = wrapper
	}

	r := rule.NewRule(ruleKind, testName)
	if wrapper != "" {
		r.AddArg(&bzl.Ident{Name: javaRuleKind})
	}

	path := strings.TrimPrefix(filepath.ToSlash(f.pathRelativeToBazelWorkspaceRoot), filepath.ToSlash(pathToPackageRelativeToBazelWorkspace+"/"))
	r.SetAttr("srcs", []string{path})
	r.SetAttr("test_class", fullyQualifiedTestClass)
	r.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{*types.NewResolvableJavaPackage(f.pkg, true, false)})

	if depOnTestHelpers != nil {
		r.SetAttr("deps", []string{*depOnTestHelpers})
	}

	if runtimeDeps.Len() != 0 {
		r.SetAttr("runtime_deps", labelsToStrings(runtimeDeps.SortedSlice()))
	}

	for k, v := range extraAttributes {
		r.SetAttr(k, v)
	}

	res.Gen = append(res.Gen, r)
	testImports := imports.Clone()
	testImports.Add(f.pkg)

	resolveInput := types.ResolveInput{
		PackageNames:         sorted_set.NewSortedSetFn([]types.PackageName{f.pkg}, types.PackageNameLess),
		ImportedPackageNames: testImports,
		ImportedClasses:      importedClasses,
		AnnotationProcessors: annotationProcessorClasses,
	}
	res.Imports = append(res.Imports, resolveInput)
}

func importsJunit4(imports *sorted_set.SortedSet[types.PackageName]) bool {
	return imports.Contains(types.NewPackageName("org.junit"))
}

// Determines whether the given import is part of the JUnit Pioneer extension pack for JUnit 5. Only the beginning of
// the string is considered here to cover classes imported from different sub-packages: org.junitpioneer.vintage.Test,
// org.junitpioneer.jupiter.RetryingTest, org.junitpioneer.jupiter.cartesian.CartesianTest, etc.
func importsJunitPioneer(import_ types.PackageName) bool {
	return strings.HasPrefix(import_.Name, "org.junitpioneer.")
}

func importsJunit5(imports *sorted_set.SortedSet[types.PackageName]) bool {
	return imports.Contains(types.NewPackageName("org.junit.jupiter.api")) ||
		imports.Contains(types.NewPackageName("org.junit.jupiter.params")) ||
		imports.Filter(importsJunitPioneer).Len() != 0
}

var junit5RuntimeDeps = []string{
	"org.junit.jupiter:junit-jupiter-engine",
	"org.junit.platform:junit-platform-launcher",
	"org.junit.platform:junit-platform-reporting",
}

func (l javaLang) generateJavaTestSuite(file *rule.File, name string, srcs []string, packageNames *sorted_set.SortedSet[types.PackageName], mavenRepositoryName string, imports *sorted_set.SortedSet[types.PackageName], importedClasses *sorted_set.SortedSet[types.ClassName], annotationProcessorClasses *sorted_set.SortedSet[types.ClassName], customTestSuffixes *[]string, hasHelpers bool, res *language.GenerateResult) {
	const ruleKind = "java_test_suite"
	r := rule.NewRule(ruleKind, name)
	r.SetAttr("srcs", srcs)
	resolvablePackages := make([]types.ResolvableJavaPackage, 0, packageNames.Len())
	if hasHelpers {
		for _, packageName := range packageNames.SortedSlice() {
			resolvablePackages = append(resolvablePackages, *types.NewResolvableJavaPackage(packageName, true, true))
		}
	}
	r.SetPrivateAttr(packagesKey, resolvablePackages)

	runtimeDeps := l.collectRuntimeDeps(ruleKind, name, file)
	if importsJunit5(imports) {
		r.SetAttr("runner", "junit5")
		for _, artifact := range junit5RuntimeDeps {
			runtimeDeps.Add(maven.LabelFromArtifact(mavenRepositoryName, artifact))
		}
		if importsJunit4(imports) {
			runtimeDeps.Add(maven.LabelFromArtifact(mavenRepositoryName, "org.junit.vintage:junit-vintage-engine"))
		}
	}

	// This should probably register imports here, and then allow the resolver to resolve this to an artifact,
	// but we don't currently wire up the resolver to do this.
	// We probably should.
	// In the mean time, hard-code some labels.
	if runtimeDeps.Len() > 0 {
		r.SetAttr("runtime_deps", labelsToStrings(runtimeDeps.SortedSlice()))
	}

	if customTestSuffixes != nil {
		r.SetAttr("test_suffixes", *customTestSuffixes)
	}

	res.Gen = append(res.Gen, r)
	suiteImports := imports.Clone()
	suiteImports.AddAll(packageNames)
	resolveInput := types.ResolveInput{
		PackageNames:         packageNames,
		ImportedPackageNames: suiteImports,
		ImportedClasses:      importedClasses,
		AnnotationProcessors: annotationProcessorClasses,
	}
	res.Imports = append(res.Imports, resolveInput)
}

// collectResourceFilesRecursively walks through subdirectories and collects resource files
func collectResourceFilesRecursively(args language.GenerateArgs, subdirPath string) []string {
	var resourceFiles []string

	// Read the subdirectory using os.ReadDir
	fullPath := filepath.Join(args.Dir, subdirPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		// If we can't read the directory, skip it
		return resourceFiles
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip BUILD files and other non-resource files
		if name == "BUILD" || name == "BUILD.bazel" {
			continue
		}

		entryPath := path.Join(subdirPath, name)

		if entry.IsDir() {
			subFiles := collectResourceFilesRecursively(args, entryPath)
			resourceFiles = append(resourceFiles, subFiles...)
		} else {
			// Check if this is a resource file
			ext := filepath.Ext(name)
			if ext != ".java" {
				resourceFiles = append(resourceFiles, entryPath)
			}
		}
	}

	return resourceFiles
}

func filterStrSlice(elts []string, f func(string) bool) []string {
	var out []string
	for _, elt := range elts {
		if !f(elt) {
			continue
		}
		out = append(out, elt)
	}
	return out
}

func labelsToStrings(labels []label.Label) []string {
	out := make([]string, len(labels))
	for idx, l := range labels {
		out[idx] = l.String()
	}
	return out
}

func testHelperLibname(targetName string) string {
	return targetName + "-test-lib"
}

func ptr[T any](v T) *T {
	return &v
}
