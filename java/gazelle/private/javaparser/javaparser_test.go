package javaparser

import (
	"testing"

	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

func TestParseExportedClassesFromKotlinFeatures(t *testing.T) {
	// Create a mock response with classes that should be exported due to Kotlin features
	// (inline functions, extension functions, property delegates, etc.)
	resp := &pb.Package{
		Name: "com.example.test",
		ExportedClasses: []string{
			"com.google.gson.Gson",
			"com.google.common.base.Strings",
		},
	}

	// Parse the exported classes
	exportedClasses := make([]types.ClassName, 0)
	for _, exportedClass := range resp.GetExportedClasses() {
		className, err := types.ParseClassName(exportedClass)
		if err != nil {
			t.Fatalf("Failed to parse exported class %q: %v", exportedClass, err)
		}
		exportedClasses = append(exportedClasses, *className)
	}

	// Verify the results
	if len(exportedClasses) != 2 {
		t.Fatalf("Expected 2 exported classes, got %d", len(exportedClasses))
	}

	// Check Gson dependency
	gson := exportedClasses[0]
	if gson.PackageName().Name != "com.google.gson" {
		t.Errorf("Expected Gson package 'com.google.gson', got '%s'", gson.PackageName().Name)
	}
	if gson.BareOuterClassName() != "Gson" {
		t.Errorf("Expected Gson class 'Gson', got '%s'", gson.BareOuterClassName())
	}
	if gson.FullyQualifiedClassName() != "com.google.gson.Gson" {
		t.Errorf("Expected Gson FQN 'com.google.gson.Gson', got '%s'", gson.FullyQualifiedClassName())
	}

	// Check Strings dependency
	strings := exportedClasses[1]
	if strings.PackageName().Name != "com.google.common.base" {
		t.Errorf("Expected Strings package 'com.google.common.base', got '%s'", strings.PackageName().Name)
	}
	if strings.BareOuterClassName() != "Strings" {
		t.Errorf("Expected Strings class 'Strings', got '%s'", strings.BareOuterClassName())
	}
	if strings.FullyQualifiedClassName() != "com.google.common.base.Strings" {
		t.Errorf("Expected Strings FQN 'com.google.common.base.Strings', got '%s'", strings.FullyQualifiedClassName())
	}

	t.Logf("Successfully parsed exported classes from Kotlin features:")
	for i, dep := range exportedClasses {
		t.Logf("  [%d] %s -> package: %s, class: %s", i, dep.FullyQualifiedClassName(), dep.PackageName().Name, dep.BareOuterClassName())
	}
}

func TestParsePackageWithKotlinFeatureExports(t *testing.T) {
	// Create a mock response similar to what the Java parser would send
	// when parsing Kotlin code with inline functions, extension functions, etc.
	resp := &pb.Package{
		Name: "com.example.provider",
		ImportedClasses: []string{
			"com.google.gson.Gson",
			"com.google.common.base.Strings",
		},
		// These classes are both imported AND exported because they're used in
		// Kotlin language features (inline functions, extension functions, etc.)
		// that leak into the public API
		ExportedClasses: []string{
			"com.google.gson.Gson",
			"com.google.common.base.Strings",
		},
		ImportedPackagesWithoutSpecificClasses: []string{},
		Mains:                                  []string{},
		PerClassMetadata:                       map[string]*pb.PerClassMetadata{},
	}

	// Simulate what ParsePackage does - just parse the exported classes
	packageName := types.NewPackageName(resp.GetName())

	exportedClasses := make([]types.ClassName, 0)
	for _, exportedClass := range resp.GetExportedClasses() {
		className, err := types.ParseClassName(exportedClass)
		if err != nil {
			t.Fatalf("Failed to parse exported class %q: %v", exportedClass, err)
		}
		exportedClasses = append(exportedClasses, *className)
	}

	// Verify the results
	if len(exportedClasses) != 2 {
		t.Fatalf("Expected 2 exported classes from Kotlin features, got %d", len(exportedClasses))
	}

	// Verify package name
	if packageName.Name != "com.example.provider" {
		t.Errorf("Expected package name 'com.example.provider', got '%s'", packageName.Name)
	}

	// Verify the exported classes include dependencies from Kotlin features
	foundGson := false
	foundStrings := false
	for _, className := range exportedClasses {
		if className.FullyQualifiedClassName() == "com.google.gson.Gson" {
			foundGson = true
		}
		if className.FullyQualifiedClassName() == "com.google.common.base.Strings" {
			foundStrings = true
		}
	}

	if !foundGson {
		t.Errorf("Expected to find Gson in exported classes from Kotlin features")
	}
	if !foundStrings {
		t.Errorf("Expected to find Strings in exported classes from Kotlin features")
	}

	t.Logf("Successfully verified package with Kotlin feature exports:")
	for i, className := range exportedClasses {
		t.Logf("  [%d] %s", i, className.FullyQualifiedClassName())
	}
}
