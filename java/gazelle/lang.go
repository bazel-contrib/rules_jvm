package gazelle

import (
	"os"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/logconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/rs/zerolog"
)

// javaLang is a language.Language implementation for Java.
type javaLang struct {
	config.Configurer
	resolve.Resolver

	parser        *javaparser.Runner
	logger        zerolog.Logger
	javaLogLevel  string
	mavenResolver maven.Resolver

	// javaPackageCache is used for module granularity support
	// Key is the path to the java package from the Bazel workspace root.
	javaPackageCache map[string]*java.Package
}

func NewLanguage() language.Language {
	goLevel, javaLevel := logconfig.LogLevel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(goLevel)
	logger.Print("creating java language")

	l := javaLang{
		logger:           logger,
		javaLogLevel:     javaLevel,
		javaPackageCache: make(map[string]*java.Package),
	}

	l.Configurer = NewConfigurer(&l)
	l.Resolver = NewResolver(&l)

	return &l
}

func (l javaLang) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"java_binary": {
			NonEmptyAttrs: map[string]bool{
				"deps": true,
				"srcs": true,
			},
			MergeableAttrs: map[string]bool{"srcs": true},
			ResolveAttrs: map[string]bool{
				"deps":         true,
				"runtime_deps": true,
			},
		},
		"java_library": {
			NonEmptyAttrs: map[string]bool{
				"deps": true,
				"srcs": true,
			},
			MergeableAttrs: map[string]bool{"srcs": true},
			ResolveAttrs: map[string]bool{
				"deps":         true,
				"runtime_deps": true,
			},
		},
		"java_test": {
			NonEmptyAttrs: map[string]bool{
				"deps": true,
				"srcs": true,
			},
			MergeableAttrs: map[string]bool{"srcs": true},
			ResolveAttrs: map[string]bool{
				"deps":         true,
				"runtime_deps": true,
			},
		},
		"java_test_suite": {
			NonEmptyAttrs: map[string]bool{
				"deps": true,
				"srcs": true,
			},
			MergeableAttrs: map[string]bool{"srcs": true},
			ResolveAttrs: map[string]bool{
				"deps":         true,
				"runtime_deps": true,
			},
		},
		"java_proto_library": {
			NonEmptyAttrs: map[string]bool{
				"deps": true,
				"srcs": true,
			},
			MergeableAttrs: map[string]bool{"srcs": true},
			ResolveAttrs: map[string]bool{
				"deps": true,
			},
		},
	}
}

var loads = []rule.LoadInfo{
	{
		Name: "@io_grpc_grpc_java//:java_grpc_library.bzl",
		Symbols: []string{
			"java_grpc_library",
		},
	},
	{
		Name: "@rules_java//java:defs.bzl",
		Symbols: []string{
			"java_binary",
			"java_library",
			"java_proto_library",
			"java_test",
		},
	},
	{
		Name: "@contrib_rules_jvm//java:defs.bzl",
		Symbols: []string{
			"java_test_suite",
		},
	},
}

func (l javaLang) Loads() []rule.LoadInfo {
	return loads
}

func (l javaLang) Fix(c *config.Config, f *rule.File) {}
