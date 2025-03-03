package gazelle

import (
	"context"
	"os"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/logconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_multiset"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/rs/zerolog"
)

// javaLang is a language.Language implementation for Java.
type javaLang struct {
	config.Configurer
	language.BaseLifecycleManager
	resolve.Resolver

	parser        *javaparser.Runner
	logger        zerolog.Logger
	javaLogLevel  string
	mavenResolver maven.Resolver

	// javaPackageCache is used for module granularity support
	// Key is the path to the java package from the Bazel workspace root.
	javaPackageCache map[string]*java.Package

	// hasHadErrors triggers the extension to fail at destroy time.
	//
	// this is used to return != 0 when some errors during the generation were
	// raised that will create invalid build files.
	hasHadErrors bool
}

func NewLanguage() language.Language {
	goLevel, javaLevel := logconfig.LogLevel()

	var logger zerolog.Logger
	if os.Getenv("GAZELLE_JAVA_LOG_FORMAT") == "json" {
		logger = zerolog.New(os.Stderr)
	} else {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if os.Getenv("GAZELLE_JAVA_LOG_TIMESTAMP") != "false" {
		logger = logger.With().Timestamp().Logger()
	}

	if os.Getenv("GAZELLE_JAVA_LOG_CALLER") != "false" {
		logger = logger.With().Caller().Logger()
	}

	logger = logger.Level(goLevel)

	logger.Debug().Msg("creating java language")

	l := javaLang{
		logger:           logger,
		javaLogLevel:     javaLevel,
		javaPackageCache: make(map[string]*java.Package),
	}

	l.logger = l.logger.Hook(shutdownServerOnFatalLogHook{
		l: &l,
	})

	l.Configurer = NewConfigurer(&l)
	l.Resolver = NewResolver(&l)

	return &l
}

var kindWithRuntimeDeps = rule.KindInfo{
	NonEmptyAttrs: map[string]bool{
		"deps": true,
		"srcs": true,
	},
	MergeableAttrs: map[string]bool{"srcs": true},
	ResolveAttrs: map[string]bool{
		"deps":         true,
		"plugins":      true,
		"runtime_deps": true,
	},
}
var kindWithoutRuntimeDeps = rule.KindInfo{
	NonEmptyAttrs: map[string]bool{
		"deps": true,
		"srcs": true,
	},
	MergeableAttrs: map[string]bool{"srcs": true},
	ResolveAttrs: map[string]bool{
		"deps":    true,
		"plugins": true,
	},
}

var javaLibraryKind = rule.KindInfo{
	NonEmptyAttrs: map[string]bool{
		"deps":    true,
		"exports": true,
		"srcs":    true,
	},
	MergeableAttrs: map[string]bool{"srcs": true},
	ResolveAttrs: map[string]bool{
		"deps":         true,
		"exports":      true,
		"plugins":      true,
		"runtime_deps": true,
	},
}

func (l javaLang) Kinds() map[string]rule.KindInfo {
	kinds := map[string]rule.KindInfo{
		"java_binary":        kindWithRuntimeDeps,
		"java_junit5_test":   kindWithRuntimeDeps,
		"java_library":       javaLibraryKind,
		"java_test":          kindWithRuntimeDeps,
		"java_test_suite":    kindWithRuntimeDeps,
		"java_proto_library": kindWithoutRuntimeDeps,
		"java_grpc_library":  kindWithoutRuntimeDeps,
	}

	c := l.Configurer.(*Configurer)
	for _, wrapper := range c.annotationToWrapper {
		kinds[wrapper.symbol] = kindWithRuntimeDeps
	}

	return kinds
}

var baseJavaLoads = []rule.LoadInfo{
	{
		Name: "@grpc-java//:java_grpc_library.bzl",
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
			"java_junit5_test",
			"java_test_suite",
		},
	},
}

func (l javaLang) Loads() []rule.LoadInfo {
	c := l.Configurer.(*Configurer)
	if len(c.annotationToWrapper) == 0 {
		return baseJavaLoads
	}

	s := sorted_multiset.NewSortedMultiSet[string, string]()
	for _, li := range baseJavaLoads {
		for _, symbol := range li.Symbols {
			s.Add(li.Name, symbol)
		}
	}

	for _, wrapper := range c.annotationToWrapper {
		s.Add(wrapper.from, wrapper.symbol)
	}

	var loads []rule.LoadInfo
	for _, name := range s.Keys() {
		loads = append(loads, rule.LoadInfo{
			Name:    name,
			Symbols: s.SortedValues(name),
		})
	}
	return loads
}

func (l javaLang) Fix(c *config.Config, f *rule.File) {}

func (l javaLang) DoneGeneratingRules() {
	l.parser.ServerManager().Shutdown()
}

func (l javaLang) AfterResolvingDeps(_ context.Context) {
	if l.hasHadErrors {
		l.logger.Fatal().Msg("the java extension encontered errors that will create invalid build files")
	}
}

type shutdownServerOnFatalLogHook struct {
	l *javaLang
}

func (s shutdownServerOnFatalLogHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	if s.l.parser == nil {
		return
	}
	if level != zerolog.FatalLevel {
		return
	}
	s.l.parser.ServerManager().Shutdown()
}
