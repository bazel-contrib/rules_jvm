package javaconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	bzl "github.com/bazelbuild/buildtools/build"
)

const (
	// JavaExcludeArtifact tells the resolver to disregard a given maven artifact.
	// Can be repeated.
	JavaExcludeArtifact = "java_exclude_artifact"

	// JavaExtensionDirective represents the directive that controls whether
	// this Java extension is enabled or not. Sub-packages inherit this value.
	// Can be either "enabled" or "disabled". Defaults to "enabled".
	JavaExtensionDirective = "java_extension"

	// JavaMavenInstallFile represents the directive that controls where the
	// maven_install.json file is located.
	// Defaults to "maven_install.json".
	JavaMavenInstallFile = "java_maven_install_file"

	// JavaModuleGranularityDirective represents the directive that controls whether
	// this Java module has a module granularity (Gradle) or a package
	// granularity (bazel).
	// Can be either "package" or "module". Defaults to "package".
	JavaModuleGranularityDirective = "java_module_granularity"

	// JavaTestFileSuffixes indicates within a test directory which files are test classes vs utility classes,
	// based on their basename.
	// It should be set up to match the value used for java_test_suite's test_suffixes attribute.
	// Accepted values are a comma-delimited list of strings.
	JavaTestFileSuffixes = "java_test_file_suffixes"

	// JavaTestMode allows user to choose from per file test or per directory test suite.
	JavaTestMode = "java_test_mode"

	// JavaGenerateProto tells the code generator whether to generate `java_proto_library` (and `java_library`)
	// rules when a `proto_library` rule is present.
	// Can be either "true" or "false". Defaults to "true".
	JavaGenerateProto = "java_generate_proto"

	// JavaMavenRepositoryName tells the code generator what the repository name that contains all maven dependencies is.
	// Defaults to "maven"
	JavaMavenRepositoryName = "java_maven_repository_name"
)

// Configs is an extension of map[string]*Config. It provides finding methods
// on top of the mapping.
type Configs map[string]*Config

// NewChild creates a new child Config. It inherits desired values from the
// current Config and sets itself as the parent to the child.
func (c *Config) NewChild() *Config {
	clonedExcludedArtifacts := make(map[string]struct{})
	for key, value := range c.excludedArtifacts {
		clonedExcludedArtifacts[key] = value
	}
	return &Config{
		parent:                 c,
		extensionEnabled:       c.extensionEnabled,
		isModuleRoot:           false,
		generateProto:          true,
		mavenInstallFile:       c.mavenInstallFile,
		moduleGranularity:      c.moduleGranularity,
		repoRoot:               c.repoRoot,
		testMode:               c.testMode,
		customTestFileSuffixes: c.customTestFileSuffixes,
		annotationToAttribute:  c.annotationToAttribute,
		annotationToWrapper:    c.annotationToWrapper,
		excludedArtifacts:      clonedExcludedArtifacts,
		mavenRepositoryName:    c.mavenRepositoryName,
	}
}

// ParentForPackage returns the parent Config for the given Bazel package.
func (c *Configs) ParentForPackage(pkg string) *Config {
	dir := filepath.Dir(pkg)
	if dir == "." {
		dir = ""
	}
	parent := (map[string]*Config)(*c)[dir]
	return parent
}

// Config represents a config extension for a specific Bazel package.
type Config struct {
	parent *Config

	extensionEnabled       bool
	isModuleRoot           bool
	generateProto          bool
	mavenInstallFile       string
	moduleGranularity      string
	repoRoot               string
	testMode               string
	customTestFileSuffixes *[]string
	excludedArtifacts      map[string]struct{}
	annotationToAttribute  map[string]map[string]bzl.Expr
	annotationToWrapper    map[string]string
	mavenRepositoryName    string
}

type LoadInfo struct {
	From   string
	Symbol string
}

// New creates a new Config.
func New(repoRoot string) *Config {
	return &Config{
		extensionEnabled:       true,
		isModuleRoot:           false,
		generateProto:          true,
		mavenInstallFile:       "maven_install.json",
		moduleGranularity:      "package",
		repoRoot:               repoRoot,
		testMode:               "suite",
		customTestFileSuffixes: nil,
		excludedArtifacts:      make(map[string]struct{}),
		annotationToAttribute:  make(map[string]map[string]bzl.Expr),
		annotationToWrapper:    make(map[string]string),
		mavenRepositoryName:    "maven",
	}
}

// ExtensionEnabled returns whether the extension is enabled or not.
func (c *Config) ExtensionEnabled() bool {
	return c.extensionEnabled
}

// SetExtensionEnabled sets whether the extension is enabled or not.
func (c *Config) SetExtensionEnabled(enabled bool) {
	c.extensionEnabled = enabled
}

func (c Config) IsModuleRoot() bool {
	return c.isModuleRoot
}

func (c *Config) GenerateProto() bool {
	return c.generateProto
}

func (c *Config) SetGenerateProto(generate bool) {
	c.generateProto = generate
}

func (c *Config) MavenRepositoryName() string {
	return c.mavenRepositoryName
}

func (c *Config) SetMavenRepositoryName(name string) {
	c.mavenRepositoryName = name
}

func (c Config) MavenInstallFile() string {
	return filepath.Join(c.repoRoot, c.mavenInstallFile)
}

func (c *Config) SetMavenInstallFile(filename string) {
	c.mavenInstallFile = filename
}

func (c Config) ModuleGranularity() string {
	return c.moduleGranularity
}

func (c *Config) SetModuleGranularity(granularity string) error {
	if granularity != "module" && granularity != "package" {
		return fmt.Errorf("%s: possible values are module/package", granularity)
	}

	if granularity == "module" {
		if c.parent == nil || c.parent.moduleGranularity == "package" {
			c.isModuleRoot = true
		}
	}

	c.moduleGranularity = granularity

	return nil
}

func (c Config) TestMode() string {
	return c.testMode
}

func (c *Config) SetTestMode(mode string) error {
	if mode != "file" && mode != "suite" {
		return fmt.Errorf("%s: possible values are 'file' or 'suite'", mode)
	}

	c.testMode = mode
	return nil
}

func (c *Config) IsJavaTestFile(basename string) bool {
	// This variable is generated by a genrule //java/gazelle/javaconfig:generate_default_java_test_patterns_src.
	suffixes := defaultTestFileSuffixes
	if c.customTestFileSuffixes != nil {
		suffixes = *c.customTestFileSuffixes
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(basename, suffix) {
			return true
		}
	}
	return false
}

func (c *Config) SetJavaTestFileSuffixes(suffixesString string) error {
	suffixes := strings.Split(suffixesString, ",")
	if equalStringSlices(suffixes, defaultTestFileSuffixes) {
		c.customTestFileSuffixes = nil
	} else {
		c.customTestFileSuffixes = &suffixes
	}
	return nil
}

func (c *Config) GetCustomJavaTestFileSuffixes() *[]string {
	return c.customTestFileSuffixes
}

func (c Config) ExcludedArtifacts() map[string]struct{} {
	return c.excludedArtifacts
}

func (c *Config) AddExcludedArtifact(s string) error {
	c.excludedArtifacts[s] = struct{}{}
	return nil
}

func (c *Config) MapAnnotationToAttribute(annotation string, key string, value bzl.Expr) {
	if _, ok := c.annotationToAttribute[annotation]; !ok {
		c.annotationToAttribute[annotation] = make(map[string]bzl.Expr)
	}
	c.annotationToAttribute[annotation][key] = value
}

func (c *Config) AttributesForAnnotation(annotation string) (map[string]bzl.Expr, bool) {
	m, ok := c.annotationToAttribute[annotation]
	return m, ok
}

func (c *Config) MapAnnotationToWrapper(annotation string, wrapper string) {
	c.annotationToWrapper[annotation] = wrapper
}

func (c *Config) WrapperForAnnotation(annotation string) (string, bool) {
	s, ok := c.annotationToWrapper[annotation]
	return s, ok
}

func (c *Config) IsTestRule(ruleKind string) bool {
	if ruleKind == "java_junit5_test" || ruleKind == "java_test" || ruleKind == "java_test_suite" {
		return true
	}
	for _, wrapper := range c.annotationToWrapper {
		if ruleKind == wrapper {
			return true
		}
	}
	return false
}

func equalStringSlices(l, r []string) bool {
	if len(l) != len(r) {
		return false
	}
	for i := range l {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}
