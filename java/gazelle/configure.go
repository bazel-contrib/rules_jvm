package gazelle

import (
	"flag"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/bazel"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// Configurer satisfies the config.Configurer interface. It's the
// language-specific configuration extension.
//
// See config.Configurer for more information.
type Configurer struct {
	lang *javaLang
}

func NewConfigurer(lang *javaLang) *Configurer {
	return &Configurer{lang}
}

func (jc Configurer) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {}

func (jc *Configurer) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	return nil
}

func (jc *Configurer) KnownDirectives() []string {
	return []string{
		javaconfig.MavenInstallFile,
		javaconfig.ModuleGranularityDirective,
		javaconfig.TestMode,
	}
}

func (jc *Configurer) Configure(c *config.Config, rel string, f *rule.File) {
	if _, exists := c.Exts[languageName]; !exists {
		outputBase, err := bazel.OutputBase(c.RepoRoot)
		if err != nil {
			jc.lang.logger.Fatal().Err(err).Msg("error getting output_base")
		}
		c.Exts[languageName] = javaconfig.Configs{
			"": javaconfig.New(c.RepoRoot, outputBase),
		}
	}

	cfgs := c.Exts[languageName].(javaconfig.Configs)
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
			}
		}
	}

	if jc.lang.parser == nil {
		jc.lang.parser = javaparser.NewRunner(jc.lang.logger, c.RepoRoot, jc.lang.javaLogLevel)
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
