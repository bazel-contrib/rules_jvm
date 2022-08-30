package gazelle

import (
	"flag"
	"fmt"
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
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
		javaconfig.JavaMavenRepositoryName,
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

func (jc *Configurer) Configure(c *config.Config, rel string, f *rule.File) {
	cfgs := jc.initRootConfig(c)
	cfg, exists := cfgs[rel]
	if !exists {
		parent := cfgs.ParentForPackage(rel)
		cfg = parent.NewChild()
		cfgs[rel] = cfg
	}

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
					jc.lang.logger.Fatal().Msgf("invalid value for directive %q: %s: possible values are true/false",
						javaconfig.JavaGenerateProto, d.Value)
				}
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
