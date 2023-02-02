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
	mavenInstallFile      string
}

func NewConfigurer(lang *javaLang) *Configurer {
	return &Configurer{
		lang:                  lang,
		annotationToAttribute: make(annotationToAttribute),
	}
}

func (jc Configurer) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	fs.Var(&jc.annotationToAttribute, "java-annotation-to-attribute", "Mapping of annotations (on test classes) to attributes which should be set for that test rule. Examples: com.example.annotations.FlakyTest=flaky=True com.example.annotations.SlowTest=timeout=\"long\"")
	fs.StringVar(&jc.mavenInstallFile, "java-maven-install-file", "", "Represents the directive that controls where the maven_install.json file is located. Defaults to \"maven_install.json\".")
}

func (jc *Configurer) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	cfgs := jc.initRootConfig(c)
	for annotation, kv := range jc.annotationToAttribute {
		for k, v := range kv {
			cfgs[""].MapAnnotationToAttribute(annotation, k, v)
		}
	}
	if jc.mavenInstallFile != "" {
		cfgs[""].SetMavenInstallFile(jc.mavenInstallFile)
	}
	return nil
}

func (jc *Configurer) KnownDirectives() []string {
	return []string{
		javaconfig.JavaExtensionDirective,
		javaconfig.MavenInstallFile,
		javaconfig.ModuleGranularityDirective,
		javaconfig.TestMode,
		javaconfig.ExcludeArtifact,
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

			case javaconfig.MavenInstallFile:
				cfg.SetMavenInstallFile(d.Value)

			case javaconfig.ModuleGranularityDirective:
				if err := cfg.SetModuleGranularity(d.Value); err != nil {
					jc.lang.logger.Fatal().Err(err).Msgf("invalid value for directive %q", javaconfig.ModuleGranularityDirective)
				}

			case javaconfig.TestMode:
				cfg.SetTestMode(d.Value)
			case javaconfig.ExcludeArtifact:
				cfg.AddExcludedArtifact(d.Value)
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
			cfg.ExcludedArtifacts(),
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
