package gazelle

import (
	"flag"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
	bzl "github.com/bazelbuild/buildtools/build"
)

// Configurer satisfies the config.Configurer interface. It's the
// language-specific configuration extension.
//
// See config.Configurer for more information.
type Configurer struct {
	lang                  *javaLang
	annotationToAttribute annotationToAttribute
	annotationToWrapper   annotationToWrapper
	mavenInstallFile      string
}

func NewConfigurer(lang *javaLang) *Configurer {
	return &Configurer{
		lang:                  lang,
		annotationToAttribute: make(annotationToAttribute),
		annotationToWrapper:   make(annotationToWrapper),
	}
}

func (jc *Configurer) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	fs.Var(&jc.annotationToAttribute, "java-annotation-to-attribute", "Mapping of annotations (on test classes) to attributes which should be set for that test rule. Examples: com.example.annotations.FlakyTest=flaky=True com.example.annotations.SlowTest=timeout=\"long\"")
	fs.Var(&jc.annotationToWrapper, "java-annotation-to-wrapper", "Mapping of annotations (on test classes) to wrapper rules which should be used around the test rule. Example: com.example.annotations.RequiresNetwork=@some//wrapper:file.bzl=requires_network")
	fs.StringVar(&jc.mavenInstallFile, "java-maven-install-file", "", "Path of the maven_install.json file. Defaults to \"maven_install.json\".")
}

func (jc *Configurer) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	cfgs := jc.initRootConfig(c)
	for annotation, kv := range jc.annotationToAttribute {
		for k, v := range kv {
			cfgs[""].MapAnnotationToAttribute(annotation, k, v)
		}
	}
	for annotation, wrapper := range jc.annotationToWrapper {
		cfgs[""].MapAnnotationToWrapper(annotation, wrapper.symbol)
	}
	if jc.mavenInstallFile != "" {
		cfgs[""].SetMavenInstallFile(jc.mavenInstallFile)
	}
	return nil
}

func (jc *Configurer) KnownDirectives() []string {
	return []string{
		javaconfig.JavaExcludeArtifact,
		javaconfig.JavaExtensionDirective,
		javaconfig.JavaMavenInstallFile,
		javaconfig.JavaModuleGranularityDirective,
		javaconfig.JavaTestFileSuffixes,
		javaconfig.JavaTestMode,
		javaconfig.JavaGenerateProto,
		javaconfig.JavaGenerateResources,
		javaconfig.JavaMavenRepositoryName,
		javaconfig.JavaAnnotationProcessorPlugin,
		javaconfig.JavaResolveToJavaExports,
		javaconfig.JavaSourcesetRoot,
		javaconfig.JavaStripResourcesPrefix,
		javaconfig.JavaGenerateBinary,
		javaconfig.JvmKotlinEnabled,
		javaconfig.JavaSearch,
		javaconfig.JavaMavenLayout,
		javaconfig.JavaSearchExclude,
	}
}

func (jc *Configurer) initRootConfig(c *config.Config) javaconfig.Configs {
	if _, exists := c.Exts[languageName]; !exists {
		c.Exts[languageName] = javaconfig.Configs{
			"": javaconfig.New(c.RepoRoot),
		}
	}
	return c.Exts[languageName].(javaconfig.Configs)
}

const binaryConfigError string = "invalid value for directive %q: %s: possible values are true/false"

func (jc *Configurer) Configure(c *config.Config, rel string, f *rule.File) {
	cfgs := jc.initRootConfig(c)
	cfg, exists := cfgs[rel]
	if !exists {
		parent := cfgs.ParentForPackage(rel)
		cfg = parent.NewChild()
		cfgs[rel] = cfg
	}

	// Auto-detect sourceset structure if not explicitly set
	if cfg.StripResourcesPrefix() == "" && cfg.SourcesetRoot() == "" {
		// Walk up the directory tree looking for sourceset patterns
		currentPath := rel
		for currentPath != "" && currentPath != "." {
			dir := filepath.Base(currentPath)
			parent := filepath.Dir(currentPath)

			// Check for Maven-style sourceset pattern: src/main/java or src/test/java
			if dir == "java" && parent != "" && parent != "." {
				grandparent := filepath.Dir(parent)
				parentBase := filepath.Base(parent)
				if grandparent != "" && filepath.Base(grandparent) == "src" {
					// Found a sourceset pattern - store the sourceset root
					// This handles src/main, src/test, src/sample, etc.
					// Use path.Join to ensure forward slashes for Bazel paths
					sourcesetRoot := path.Join(grandparent, parentBase)
					cfg.SetSourcesetRoot(sourcesetRoot)
					// Also set the strip prefix for resources
					resourcesRoot := path.Join(sourcesetRoot, "resources")
					cfg.SetStripResourcesPrefix(resourcesRoot)
					break
				}
			}

			// Also check if we're in a resources directory
			if dir == "resources" && parent != "" && parent != "." {
				grandparent := filepath.Dir(parent)
				parentBase := filepath.Base(parent)
				if grandparent != "" && filepath.Base(grandparent) == "src" {
					// Found a sourceset pattern from resources side
					sourcesetRoot := path.Join(grandparent, parentBase)
					cfg.SetSourcesetRoot(sourcesetRoot)
					// Also set the strip prefix for resources
					resourcesRoot := path.Join(sourcesetRoot, "resources")
					cfg.SetStripResourcesPrefix(resourcesRoot)
					break
				}
			}

			currentPath = parent
		}
	}

	// Process directives from BUILD file
	if f != nil {
		for _, d := range f.Directives {
			switch d.Key {
			case javaconfig.JavaExcludeArtifact:
				cfg.AddExcludedArtifact(d.Value)

			case javaconfig.JavaExtensionDirective:
				switch d.Value {
				case "enabled":
					cfg.SetExtensionEnabled(true)
				case "disabled":
					cfg.SetExtensionEnabled(false)
				default:
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: %s: possible values are enabled/disabled",
						javaconfig.JavaExtensionDirective, d.Value)
				}

			case javaconfig.JavaMavenInstallFile:
				cfg.SetMavenInstallFile(d.Value)

			case javaconfig.JavaModuleGranularityDirective:
				if err := cfg.SetModuleGranularity(d.Value); err != nil {
					jc.lang.logger.Fatal().Err(err).Msgf("invalid value for directive %q", javaconfig.JavaModuleGranularityDirective)
				}

			case javaconfig.JavaTestFileSuffixes:
				cfg.SetJavaTestFileSuffixes(d.Value)

			case javaconfig.JavaTestMode:
				cfg.SetTestMode(d.Value)

			case javaconfig.JavaMavenRepositoryName:
				cfg.SetMavenRepositoryName(d.Value)

			case javaconfig.JavaGenerateProto:
				switch d.Value {
				case "true":
					cfg.SetGenerateProto(true)
				case "false":
					cfg.SetGenerateProto(false)
				default:
					jc.lang.logger.Fatal().Msgf(binaryConfigError, javaconfig.JavaGenerateProto, d.Value)
				}
			case javaconfig.JavaGenerateBinary:
				switch d.Value {
				case "true":
					cfg.SetGenerateBinary(true)
				case "false":
					cfg.SetGenerateBinary(false)
				default:
					jc.lang.logger.Fatal().Msgf(binaryConfigError, javaconfig.JavaGenerateBinary, d.Value)
				}
			case javaconfig.JavaGenerateResources:
				switch d.Value {
				case "true":
					cfg.SetGenerateResources(true)
				case "false":
					cfg.SetGenerateResources(false)
				default:
					jc.lang.logger.Fatal().Msgf(binaryConfigError, javaconfig.JavaGenerateResources, d.Value)
				}
			case javaconfig.JavaAnnotationProcessorPlugin:
				// Format: # gazelle:java_annotation_processor_plugin com.example.AnnotationName com.example.AnnotationProcessorImpl
				parts := strings.Split(d.Value, " ")
				if len(parts) != 2 {
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: %s: expected an annotation class-name followed by a processor class-name", javaconfig.JavaAnnotationProcessorPlugin, d.Value)
				}
				annotationClassName, err := types.ParseClassName(parts[0])
				if err != nil {
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: %q: couldn't parse annotation processor annotation class-name: %v", javaconfig.JavaAnnotationProcessorPlugin, parts[0], err)
				}
				processorClassName, err := types.ParseClassName(parts[1])
				if err != nil {
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: %q: couldn't parse annotation processor class-name: %v", javaconfig.JavaAnnotationProcessorPlugin, parts[1], err)
				}
				cfg.AddAnnotationProcessorPlugin(*annotationClassName, *processorClassName)

			case javaconfig.JavaResolveToJavaExports:
				if !cfg.CanSetResolveToJavaExports() {
					jc.lang.logger.Fatal().
						Msgf("Detected multiple attempts to initialize directive %q. Please only initialize it once for the entire repository.",
							javaconfig.JavaResolveToJavaExports)
				}
				if rel != "" {
					jc.lang.logger.Fatal().
						Msgf("Enabling or disabling directive %q must be done from the root of the repository.",
							javaconfig.JavaResolveToJavaExports)
				}
				switch d.Value {
				case "true":
					cfg.SetResolveToJavaExports(true)
				case "false":
					cfg.SetResolveToJavaExports(false)
				default:
					jc.lang.logger.Fatal().Msgf(binaryConfigError, javaconfig.JavaResolveToJavaExports, d.Value)
				}

			case javaconfig.JavaSourcesetRoot:
				cfg.SetSourcesetRoot(d.Value)

			case javaconfig.JavaStripResourcesPrefix:
				cfg.SetStripResourcesPrefix(d.Value)

			case javaconfig.JvmKotlinEnabled:
				switch d.Value {
				case "true":
					cfg.SetKotlinEnabled(true)
				case "false":
					cfg.SetKotlinEnabled(false)
				default:
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: %s: possible values are true/false",
						javaconfig.JvmKotlinEnabled, d.Value)
				}

			case javaconfig.JavaSearch:
				if err := cfg.AddSearchPath(d.Value); err != nil {
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: %s: %v",
						javaconfig.JavaSearch, d.Value, err)
				}

			case javaconfig.JavaMavenLayout:
				discovered, err := cfg.DiscoverMavenLayout(d.Value)
				if err != nil {
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: %s: %v",
						javaconfig.JavaMavenLayout, d.Value, err)
				}
				if len(discovered) > 0 {
					jc.lang.logger.Debug().Strs("paths", discovered).Msg("discovered maven layout source roots")
				}

			case javaconfig.JavaSearchExclude:
				// Format: <directory>
				if d.Value == "" {
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: expected a directory path",
						javaconfig.JavaSearchExclude)
				}
				cfg.AddSearchExclude(d.Value)
			}
		}
	}

	if jc.lang.parser == nil {
		runner, err := javaparser.NewRunner(jc.lang.logger, c.RepoRoot, jc.lang.javaLogLevel)
		if err != nil {
			jc.lang.logger.Fatal().Err(err).Msg("could not start javaparser")
		}
		jc.lang.parser = runner
	}

	if jc.lang.mavenResolver == nil {
		resolver, err := maven.NewResolver(
			cfg.MavenInstallFile(),
			jc.lang.logger,
		)
		if err != nil {
			jc.lang.logger.Fatal().Err(err).Msg("error creating Maven resolver")
		}
		jc.lang.mavenResolver = resolver
	}
}

type annotationToAttribute map[string]map[string]bzl.Expr

func (f *annotationToAttribute) String() string {
	s := "annotationToAttribute{"
	firstAnnotation := true
	for annotation, v := range *f {
		if !firstAnnotation {
			s += "\n"
		}
		s += annotation + ": "
		firstAttr := true
		for attr, val := range v {
			if !firstAttr {
				s += ", "
			}
			s += fmt.Sprintf("%s=%v", attr, val)
			firstAttr = false
		}
		firstAnnotation = false
	}
	s += "}"
	return s
}

func (f *annotationToAttribute) Set(value string) error {
	annotationToKeyToValue := strings.Split(value, "=")
	if len(annotationToKeyToValue) != 3 {
		return fmt.Errorf("want --annotation-to-attribute flag to have format com.example.Annotation=flaky=True but didn't find exactly one equals sign")
	}
	annotationClassName := annotationToKeyToValue[0]

	if _, ok := (*f)[annotationClassName]; !ok {
		(*f)[annotationClassName] = make(map[string]bzl.Expr)
	}

	key := annotationToKeyToValue[1]
	parsedValue := &bzl.LiteralExpr{Token: annotationToKeyToValue[2]}

	if existingValue, ok := (*f)[annotationClassName][key]; ok {
		return fmt.Errorf("saw duplicate values for --annotation-to-attribute flag for key %v: %v and %v", key, existingValue, parsedValue)
	}

	(*f)[annotationClassName][key] = parsedValue
	return nil
}

type loadInfo struct {
	from   string
	symbol string
}

type annotationToWrapper map[string]loadInfo

func (f *annotationToWrapper) String() string {
	s := "annotationToWrapper{"
	for a, li := range *f {
		s += a + ": "
		s += fmt.Sprintf(`load("%s", "%s")`, li.from, li.symbol)
	}
	s += "}"
	return s
}

func (f *annotationToWrapper) Set(value string) error {
	parts := strings.Split(value, "=")
	if len(parts) != 2 {
		return fmt.Errorf("want --java-annotation-to-wrapper to have format com.example.RequiresNetwork=@some_repo//has:wrapper.bzl,wrapper_rule but didn't see exactly one equals sign")
	}
	annotation := parts[0]

	if _, ok := (*f)[annotation]; ok {
		return fmt.Errorf("saw conflicting values for --java-annotation-to-wrapper flag for annotation %v", annotation)
	}

	vParts := strings.Split(parts[1], ",")
	if len(vParts) != 2 {
		return fmt.Errorf("want --java-annotation-to-wrapper to have format com.example.RequiresNetwork=@some_repo//has:wrapper.bzl,wrapper_rule but didn't see exactly one comma after equals sign")
	}

	from := vParts[0]
	symbol := vParts[1]

	(*f)[annotation] = loadInfo{
		from:   from,
		symbol: symbol,
	}

	return nil
}
