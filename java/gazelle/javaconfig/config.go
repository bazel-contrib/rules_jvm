package javaconfig

import (
	"fmt"
	"path/filepath"
)

const (
	// JavaExtensionDirective represents the directive that controls whether
	// this Java extension is enabled or not. Sub-packages inherit this value.
	// Can be either "enabled" or "disabled". Defaults to "enabled".
	JavaExtensionDirective = "java_extension"

	// ModuleGranularityDirective represents the directive that controls whether
	// this Java module has a module granularity (Gradle) or a package
	// granularity (bazel).
	// Can be either "package" or "module". Defaults to "package".
	ModuleGranularityDirective = "java_module_granularity"

	// MavenInstallFile represents the directive that controls where the
	// maven_install.json file is located.
	// Defaults to "maven_install.json".
	MavenInstallFile = "java_maven_install_file"

	// TestMode allows user to choose from per file test or per directory test suite.
	TestMode = "java_test_mode"
)

// Configs is an extension of map[string]*Config. It provides finding methods
// on top of the mapping.
type Configs map[string]*Config

// NewChild creates a new child Config. It inherits desired values from the
// current Config and sets itself as the parent to the child.
func (c *Config) NewChild() *Config {
	return &Config{
		parent:            c,
		extensionEnabled:  c.extensionEnabled,
		isModuleRoot:      false,
		mavenInstallFile:  c.mavenInstallFile,
		moduleGranularity: c.moduleGranularity,
		outputBase:        c.outputBase,
		repoRoot:          c.repoRoot,
		testMode:          c.testMode,
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

	extensionEnabled  bool
	isModuleRoot      bool
	mavenInstallFile  string
	moduleGranularity string
	outputBase        string
	repoRoot          string
	testMode          string
}

// New creates a new Config.
func New(repoRoot, outputBase string) *Config {
	return &Config{
		extensionEnabled:  true,
		isModuleRoot:      false,
		mavenInstallFile:  "maven_install.json",
		moduleGranularity: "package",
		outputBase:        outputBase,
		repoRoot:          repoRoot,
		testMode:          "suite",
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

func (c Config) OutputBase() string {
	return c.outputBase
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
