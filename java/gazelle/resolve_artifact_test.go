package gazelle

import (
	"reflect"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func TestExactClassUsesPublicExportAcrossArtifacts(t *testing.T) {
	c, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	file := rule.EmptyFile("library/BUILD.bazel", "library")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("exports", []string{
		"//library/src/main/java/com/example:provider",
		"//library/custom:consumer",
	})
	lang.javaExportIndex.RecordJavaExport(export, file)
	lang.javaExportIndex.FinalizeIndex()

	pkg := types.NewPackageName("com.example")
	class := types.NewClassName(pkg, "Provider")
	provider := label.New("", "library/src/main/java/com/example", "provider")
	resolver := lang.Resolver.(*Resolver)
	resolver.classIndex[pkg] = &packageClassIndex{
		prod: map[string][]label.Label{"Provider": {provider}},
		test: map[string][]label.Label{},
	}
	pc := javaconfig.New(".")
	pc.SetResolveToJavaExports(true)
	for _, tc := range []struct {
		name string
		from label.Label
		want label.Label
	}{
		{"external consumer", label.New("", "consumer", "consumer"), label.New("", "library", "published")},
		{"local consumer", label.New("", "library/src/test/java/com/example", "test"), provider},
		{"published sibling without src layout", label.New("", "library/custom", "consumer"), provider},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolver.resolveSingleClass(c, pc, class, nil, tc.from, false); got != tc.want {
				t.Errorf("resolveSingleClass() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestResolvedDepsRemoveCoveredLibraryAndOwningExport(t *testing.T) {
	c, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	file := rule.EmptyFile("library/BUILD.bazel", "library")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("exports", []string{"//library/internal:provider", "//library/internal:sibling"})
	export.SetAttr("maven_coordinates", "com.example:library:1.0")
	lang.javaExportIndex.RecordJavaExport(export, file)
	lang.javaExportIndex.FinalizeIndex()
	resolver := lang.Resolver.(*Resolver)
	pc := javaconfig.New(".")
	pc.SetResolveToJavaExports(true)
	publicExport := label.New("", "library", "published")

	consumer := rule.NewRule("java_library", "consumer")
	consumer.SetAttr("deps", []string{"//library/internal:provider", "//generated:proto"})
	resolver.setResolvedLabelAttrIncludingExistingValues(c, pc, consumer, "deps",
		sortedLabels(publicExport), label.New("", "consumer", "consumer"))
	if got, want := consumer.AttrStrings("deps"), []string{"//generated:proto", "//library:published"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("cross-artifact deps = %v, want %v", got, want)
	}

	provider := rule.NewRule("java_library", "provider")
	provider.SetAttr("deps", []string{"//library:published", "//library/internal:sibling", "@maven//:com_example_library"})
	providerConfig := resolver.configWithoutPublishedArtifact(pc, label.New("", "library/internal", "provider"))
	resolver.setResolvedLabelAttrIncludingExistingValues(c, providerConfig, provider, "deps",
		sortedLabels(), label.New("", "library/internal", "provider"))
	if got, want := provider.AttrStrings("deps"), []string{"//library/internal:sibling"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("own export removed the sibling library: deps = %v, want %v", got, want)
	}
}

func TestResolvedDepsKeepExternalLabelNamedLikeOwningExport(t *testing.T) {
	c, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	file := rule.EmptyFile("library/BUILD.bazel", "library")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("exports", []string{"//library/internal:provider"})
	lang.javaExportIndex.RecordJavaExport(export, file)
	lang.javaExportIndex.FinalizeIndex()
	resolver := lang.Resolver.(*Resolver)
	pc := javaconfig.New(".")
	pc.SetResolveToJavaExports(true)
	r := rule.NewRule("java_library", "provider")
	r.SetAttr("deps", []string{"@external//library:published"})
	resolver.setResolvedLabelAttrIncludingExistingValues(c, pc, r, "deps", sortedLabels(), label.New("", "library/internal", "provider"))
	if got, want := r.AttrStrings("deps"), []string{"@external//library:published"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("deps = %v, want %v", got, want)
	}
}

func TestExternalCandidateNamedLikeOwnExportIsNotDiscarded(t *testing.T) {
	_, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	file := rule.EmptyFile("library/BUILD.bazel", "library")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("exports", []string{"//library/internal:provider"})
	lang.javaExportIndex.RecordJavaExport(export, file)
	lang.javaExportIndex.FinalizeIndex()
	from := label.New("", "library/internal", "provider")
	external := resolve.FindResult{Label: label.New("external", "library", "published")}
	got := lang.Resolver.(*Resolver).tryResolvingToJavaExport([]resolve.FindResult{external}, from)
	if !reflect.DeepEqual(got, []resolve.FindResult{external}) {
		t.Fatalf("external candidate = %v, want %v", got, external)
	}
}

func sortedLabels(values ...label.Label) *sorted_set.SortedSet[label.Label] {
	return sorted_set.NewSortedSetFn(values, sorted_set.LabelLess)
}

func TestSourceArtifactRootPreservesNestedModule(t *testing.T) {
	if got, ok := sourceArtifactRoot("src/first/src/main/java"); !ok || got != "src/first" {
		t.Fatalf("sourceArtifactRoot() = %q, %t, want %q, true", got, ok, "src/first")
	}
}

func TestClassOverrideKeepsExactWorkspaceProviderAheadOfMaven(t *testing.T) {
	c, langs, configurers := testConfig(t)
	directives, err := rule.LoadData("BUILD.bazel", "", []byte("# gazelle:resolve java com.example.Overridden //generated:overridden\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, configurer := range configurers {
		configurer.Configure(c, "", directives)
	}
	lang := langs[1].(*javaLang)
	pkg := types.NewPackageName("com.example")
	local := types.NewClassName(pkg, "Local")
	lang.javaExportIndex.FinalizeIndex()
	provider := label.New("", "library", "provider")
	resolver := lang.Resolver.(*Resolver)
	resolver.classIndex[pkg] = &packageClassIndex{
		prod: map[string][]label.Label{"Local": {provider}},
		test: map[string][]label.Label{},
	}
	resolvers, extensions := InitTestResolversAndExtensions(langs)
	lang.mavenResolver = &artifactTestMavenResolver{}
	index := resolve.NewRuleIndex(resolvers.Resolver, extensions...)
	index.Finish()
	consumer := rule.NewRule("java_library", "consumer")
	resolver.populateAttr(c, javaconfig.New("."), consumer, "deps",
		testPackageNames(pkg), testClassNames(local, types.NewClassName(pkg, "MavenOnly"), types.NewClassName(pkg, "Overridden")),
		index, false, label.New("", "consumer", "consumer"), testPackageNames())
	want := []string{"//generated:overridden", "//library:provider", "@maven//:external"}
	if got := consumer.AttrStrings("deps"); !reflect.DeepEqual(got, want) {
		t.Fatalf("deps = %v, want %v", got, want)
	}
}

type artifactTestMavenResolver struct{}

func (*artifactTestMavenResolver) Resolve(types.PackageName, map[string]struct{}, string) (label.Label, error) {
	return label.NoLabel, nil
}

func (*artifactTestMavenResolver) ResolveClass(class types.ClassName, _ map[string]struct{}, _ string) (label.Label, error) {
	if class.BareOuterClassName() == "Overridden" {
		return label.New("maven", "", "overridden"), nil
	}
	return label.New("maven", "", "external"), nil
}

func TestSplitExportPackageFallsBackToClassResolution(t *testing.T) {
	_, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	for _, name := range []string{"first", "second"} {
		file := rule.EmptyFile(name+"/BUILD.bazel", name)
		export := rule.NewRule("java_export", "published")
		export.SetAttr("exports", []string{"//" + name + "/src/main/java/com/example:provider"})
		lang.javaExportIndex.RecordJavaExport(export, file)
	}
	lang.javaExportIndex.FinalizeIndex()
	results := []resolve.FindResult{
		{Label: label.New("", "first", "published")},
		{Label: label.New("", "second", "published")},
	}
	got := lang.Resolver.(*Resolver).tryResolvingToJavaExport(results, label.New("", "consumer", "consumer"))
	if !reflect.DeepEqual(got, results) {
		t.Fatalf("split package candidates = %v, want %v", got, results)
	}
}

func TestOwnExportIsAvailableToUnpublishedTestButNotPublishedLibrary(t *testing.T) {
	_, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	file := rule.EmptyFile("library/BUILD.bazel", "library")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("exports", []string{"//library/src/main/java/com/example:provider"})
	lang.javaExportIndex.RecordJavaExport(export, file)
	lang.javaExportIndex.FinalizeIndex()
	public := resolve.FindResult{Label: label.New("", "library", "published")}
	resolver := lang.Resolver.(*Resolver)
	test := label.New("", "library/src/test/java/com/example", "test")
	if got := resolver.tryResolvingToJavaExport([]resolve.FindResult{public}, test); !reflect.DeepEqual(got, []resolve.FindResult{public}) {
		t.Fatalf("test dependency = %v, want own export", got)
	}
	provider := label.New("", "library/src/main/java/com/example", "provider")
	if got := resolver.tryResolvingToJavaExport([]resolve.FindResult{public}, provider); len(got) != 0 {
		t.Fatalf("published library depends on own export: %v", got)
	}
}

func TestExportDoesNotDeclareRuntimeOnlyClass(t *testing.T) {
	_, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	file := rule.EmptyFile("library/BUILD.bazel", "library")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("runtime_deps", []string{":runtime"})
	lang.javaExportIndex.RecordJavaExport(export, file)
	class := types.NewClassName(types.NewPackageName("com.example"), "Runtime")
	lang.classExportCache[label.New("", "library", "runtime").String()] = classExportInfo{classes: []types.ClassName{class}}
	if lang.Resolver.(*Resolver).ruleDeclaresClass(label.New("", "library", "published"), class) {
		t.Fatal("runtime-only dependency was treated as a compile-time class")
	}
}

func TestPublishedLibraryExcludesOwnMavenArtifact(t *testing.T) {
	_, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	file := rule.EmptyFile("library/BUILD.bazel", "library")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("exports", []string{"//library/internal:provider"})
	export.SetAttr("maven_coordinates", "com.example:library:1.0")
	lang.javaExportIndex.RecordJavaExport(export, file)
	lang.javaExportIndex.FinalizeIndex()
	pc := javaconfig.New(".")
	from := label.New("", "library/internal", "provider")
	filtered := lang.Resolver.(*Resolver).configWithoutPublishedArtifact(pc, from)
	artifact := maven.LabelFromArtifact(pc.MavenRepositoryName(), "com.example:library").String()
	if _, found := filtered.ExcludedArtifacts()[artifact]; !found {
		t.Fatalf("own published artifact %s was not excluded", artifact)
	}
	if _, found := pc.ExcludedArtifacts()[artifact]; found {
		t.Fatal("excluding the published artifact mutated shared config")
	}
}

func TestExcludedArtifactBypassesCachedPackageResolution(t *testing.T) {
	c, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	resolvers, extensions := InitTestResolversAndExtensions(langs)
	lang.mavenResolver = &exclusionAwareMavenResolver{}
	index := resolve.NewRuleIndex(resolvers.Resolver, extensions...)
	index.Finish()
	resolver := lang.Resolver.(*Resolver)
	pc := javaconfig.New(".")
	pkg := types.NewPackageName("com.example")
	from := label.New("", "library", "consumer")
	if got := resolver.resolveSinglePackage(c, pc, pkg, index, from, false, testPackageNames(), nil); got.Name != "own" {
		t.Fatalf("initial package resolution = %s, want own artefact", got)
	}
	child := pc.NewChild()
	_ = child.AddExcludedArtifact("@maven//:own")
	if got := resolver.resolveSinglePackage(c, child, pkg, index, from, false, testPackageNames(), nil); got.Name != "other" {
		t.Fatalf("excluded package resolution = %s, want other artefact", got)
	}
}

type exclusionAwareMavenResolver struct{}

func (*exclusionAwareMavenResolver) Resolve(_ types.PackageName, excluded map[string]struct{}, _ string) (label.Label, error) {
	if _, found := excluded["@maven//:own"]; found {
		return label.New("maven", "", "other"), nil
	}
	return label.New("maven", "", "own"), nil
}

func (*exclusionAwareMavenResolver) ResolveClass(types.ClassName, map[string]struct{}, string) (label.Label, error) {
	return label.NoLabel, nil
}

func TestKotlinTestAssociatesUseArtifactLocalSplitPackageProvider(t *testing.T) {
	c, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	resolvers, extensions := InitTestResolversAndExtensions(langs)
	index := resolve.NewRuleIndex(resolvers.Resolver, extensions...)
	pkg := types.NewPackageName("com.example.shared")
	for _, artifact := range []string{"first", "second"} {
		dir := artifact + "/src/main/kotlin/com/example/shared"
		file := rule.EmptyFile(dir+"/BUILD.bazel", dir)
		library := rule.NewRule("kt_jvm_library", "library")
		library.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{*types.NewResolvableJavaPackage(pkg, false, false)})
		index.AddRule(c, library, file)
	}
	index.Finish()
	lang.javaExportIndex.FinalizeIndex()
	test := rule.NewRule("kt_jvm_library", "test")
	test.SetAttr("srcs", []string{"Test.kt"})
	input := types.ResolveInput{PackageNames: testPackageNames(pkg)}
	from := label.New("", "first/src/test/kotlin/com/example/shared", "test")
	lang.Resolver.(*Resolver).populateAssociatesAttr(c, index, input, test, true, from)
	want := []string{"//first/src/main/kotlin/com/example/shared:library"}
	if got := test.AttrStrings("associates"); !reflect.DeepEqual(got, want) {
		t.Fatalf("associates = %v, want %v", got, want)
	}
}

func TestKotlinAssociateRemovesDuplicatePublishedDeps(t *testing.T) {
	c, langs, _ := testConfig(t)
	lang := langs[1].(*javaLang)
	file := rule.EmptyFile("library/BUILD.bazel", "library")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("exports", []string{"//library/src/main/kotlin/com/example:main"})
	export.SetAttr("maven_coordinates", "com.example:library:1.0")
	lang.javaExportIndex.RecordJavaExport(export, file)
	lang.javaExportIndex.FinalizeIndex()
	r := rule.NewRule("kt_jvm_library", "test")
	r.SetAttr("associates", []string{"//library/src/main/kotlin/com/example:main"})
	r.SetAttr("deps", []string{
		"//library:published",
		"@maven//:com_example_library",
		"//library/src/main/kotlin/com/example:main",
		"@maven//:other",
	})
	from := label.New("", "library/src/test/kotlin/com/example", "test")
	lang.Resolver.(*Resolver).removeRedundantAssociateDeps(c, javaconfig.New("."), r, from)
	if got, want := r.AttrStrings("deps"), []string{"@maven//:other"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("deps = %v, want %v", got, want)
	}
}
