package javaparser

import (
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/rs/zerolog"
)

func TestParseImplicitDependencies(t *testing.T) {
	// Create a mock response with implicit dependencies
	resp := &pb.Package{
		Name: "com.example.test",
		ImplicitDeps: []string{
			"com.google.gson.Gson",
			"com.google.common.base.Strings",
		},
	}

	// Test parsing the implicit dependencies
	implicitDeps := make([]types.ClassName, 0)
	for _, depClass := range resp.GetImplicitDeps() {
		className, err := types.ParseClassName(depClass)
		if err != nil {
			t.Fatalf("Failed to parse implicit dependency %q: %v", depClass, err)
		}
		implicitDeps = append(implicitDeps, *className)
	}

	// Verify the results
	if len(implicitDeps) != 2 {
		t.Fatalf("Expected 2 implicit dependencies, got %d", len(implicitDeps))
	}

	// Check Gson dependency
	gson := implicitDeps[0]
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
	strings := implicitDeps[1]
	if strings.PackageName().Name != "com.google.common.base" {
		t.Errorf("Expected Strings package 'com.google.common.base', got '%s'", strings.PackageName().Name)
	}
	if strings.BareOuterClassName() != "Strings" {
		t.Errorf("Expected Strings class 'Strings', got '%s'", strings.BareOuterClassName())
	}
	if strings.FullyQualifiedClassName() != "com.google.common.base.Strings" {
		t.Errorf("Expected Strings FQN 'com.google.common.base.Strings', got '%s'", strings.FullyQualifiedClassName())
	}

	t.Logf("Successfully parsed implicit dependencies:")
	for i, dep := range implicitDeps {
		t.Logf("  [%d] %s -> package: %s, class: %s", i, dep.FullyQualifiedClassName(), dep.PackageName().Name, dep.BareOuterClassName())
	}
}

func TestParsePackageWithImplicitDeps(t *testing.T) {
	// Create a mock response similar to what the Java parser would send
	resp := &pb.Package{
		Name: "com.example.provider",
		ImportedClasses: []string{
			"com.google.gson.Gson",
			"com.google.common.base.Strings",
		},
		ExportedClasses:                        []string{},
		ImportedPackagesWithoutSpecificClasses: []string{},
		Mains:                                  []string{},
		PerClassMetadata:                       map[string]*pb.PerClassMetadata{},
		ImplicitDeps: []string{
			"com.google.gson.Gson",
			"com.google.common.base.Strings",
		},
	}

	// Create a test logger
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()

	// Simulate the parsing logic from ParsePackage
	packageName := types.NewPackageName(resp.GetName())

	// Parse implicit dependencies
	var implicitDeps []types.ClassName
	logger.Debug().
		Int("implicit_deps_count", len(resp.GetImplicitDeps())).
		Strs("implicit_deps_raw", resp.GetImplicitDeps()).
		Msg("Parsing implicit dependencies from Java response")

	for i, depClass := range resp.GetImplicitDeps() {
		logger.Debug().
			Int("index", i).
			Str("dependency", depClass).
			Msg("Parsing implicit dependency from Java response")

		className, err := types.ParseClassName(depClass)
		if err != nil {
			logger.Error().
				Str("dependency", depClass).
				Err(err).
				Msg("Failed to parse implicit dependency class name")
			t.Fatalf("Failed to parse implicit dependency %q: %v", depClass, err)
		}

		logger.Debug().
			Str("dependency", depClass).
			Str("parsed_package", className.PackageName().Name).
			Str("parsed_class", className.BareOuterClassName()).
			Msg("Successfully parsed implicit dependency")

		implicitDeps = append(implicitDeps, *className)
	}

	logger.Debug().
		Int("final_implicit_deps_count", len(implicitDeps)).
		Msg("Finished parsing implicit dependencies")

	// Create the java.Package (this is what would be returned)
	javaPkg := &java.Package{
		Name:         packageName,
		ImplicitDeps: implicitDeps,
	}

	// Verify the results
	if len(javaPkg.ImplicitDeps) != 2 {
		t.Fatalf("Expected 2 implicit dependencies, got %d", len(javaPkg.ImplicitDeps))
	}

	t.Logf("Successfully created java.Package with implicit dependencies:")
	for i, dep := range javaPkg.ImplicitDeps {
		t.Logf("  [%d] %s", i, dep.FullyQualifiedClassName())
	}
}
