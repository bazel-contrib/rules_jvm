package gazelle

import (
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/testtools"
	"github.com/stretchr/testify/require"
)

func TestFlagParsing(t *testing.T) {
	configurer := NewConfigurer(NewLanguage().(*javaLang))

	gazelleConfig := testtools.NewTestConfig(t,
		[]config.Configurer{configurer},
		[]language.Language{},
		[]string{
			"-java-annotation-to-attribute=com.example.annotations.FlakyTest=flaky=True",
			"-java-maven-install-file=install_maven.json",
		})

	// Command line value made it to the configurer
	require.Equal(t, "install_maven.json", configurer.mavenInstallFile)

	// Command line value made it to the java config
	javaConfig := gazelleConfig.Exts[languageName].(javaconfig.Configs)
	require.Equal(t, "install_maven.json", javaConfig[""].MavenInstallFile())
}
