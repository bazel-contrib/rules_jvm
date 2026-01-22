package maven

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_loadConfiguration_v1(t *testing.T) {
	cfg, err := loadConfiguration("testdata/v1_maven_install.json")
	require.NoError(t, err)

	require.ElementsMatch(t, cfg.ListDependencies(), []string{
		"com.google.code.findbugs:jsr305:3.0.2",
		"com.google.errorprone:error_prone_annotations:2.11.0",
		"com.google.guava:failureaccess:1.0.1",
		"com.google.guava:guava:31.1-jre",
		"com.google.guava:listenablefuture:9999.0-empty-to-avoid-conflict-with-guava",
		"com.google.j2objc:j2objc-annotations:1.3",
		"org.checkerframework:checker-qual:3.12.0",
	})

	require.Equal(t, cfg.GetDependencyCoordinates("com.google.guava:guava:31.1-jre"), "com.google.guava:guava:31.1-jre")

	require.ElementsMatch(t, cfg.ListDependencyPackages("com.google.guava:guava:31.1-jre"), []string{
		"com.google.common.annotations",
		"com.google.common.base.internal",
		"com.google.common.base",
		"com.google.common.cache",
		"com.google.common.collect",
		"com.google.common.escape",
		"com.google.common.eventbus",
		"com.google.common.graph",
		"com.google.common.hash",
		"com.google.common.html",
		"com.google.common.io",
		"com.google.common.math",
		"com.google.common.net",
		"com.google.common.primitives",
		"com.google.common.reflect",
		"com.google.common.util.concurrent",
		"com.google.common.xml",
		"com.google.thirdparty.publicsuffix",
	})
}

func Test_loadConfiguration_v2(t *testing.T) {
	cfg, err := loadConfiguration("testdata/v2_maven_install.json")
	require.NoError(t, err)

	require.ElementsMatch(t, cfg.ListDependencies(), []string{
		"com.google.code.findbugs:jsr305",
		"com.google.errorprone:error_prone_annotations",
		"com.google.guava:failureaccess",
		"com.google.guava:guava",
		"com.google.guava:listenablefuture",
		"com.google.j2objc:j2objc-annotations",
		"org.checkerframework:checker-qual",
	})

	require.Equal(t, cfg.GetDependencyCoordinates("com.google.guava:guava"), "com.google.guava:guava:31.1-jre")

	require.ElementsMatch(t, cfg.ListDependencyPackages("com.google.guava:guava"), []string{
		"com.google.common.annotations",
		"com.google.common.base.internal",
		"com.google.common.base",
		"com.google.common.cache",
		"com.google.common.collect",
		"com.google.common.escape",
		"com.google.common.eventbus",
		"com.google.common.graph",
		"com.google.common.hash",
		"com.google.common.html",
		"com.google.common.io",
		"com.google.common.math",
		"com.google.common.net",
		"com.google.common.primitives",
		"com.google.common.reflect",
		"com.google.common.util.concurrent",
		"com.google.common.xml",
		"com.google.thirdparty.publicsuffix",
	})
}

func Test_loadConfiguration_v3(t *testing.T) {
	cfg, err := loadConfiguration("testdata/v3_maven_install.json")
	require.NoError(t, err)

	require.ElementsMatch(t, cfg.ListDependencies(), []string{
		"com.google.code.findbugs:jsr305",
		"com.google.errorprone:error_prone_annotations",
		"com.google.guava:failureaccess",
		"com.google.guava:guava",
		"com.google.guava:listenablefuture",
		"com.google.j2objc:j2objc-annotations",
		"org.checkerframework:checker-qual",
	})

	require.Equal(t, cfg.GetDependencyCoordinates("com.google.guava:guava"), "com.google.guava:guava:31.1-jre")

	require.ElementsMatch(t, cfg.ListDependencyPackages("com.google.guava:guava"), []string{
		"com.google.common.annotations",
		"com.google.common.base.internal",
		"com.google.common.base",
		"com.google.common.cache",
		"com.google.common.collect",
		"com.google.common.escape",
		"com.google.common.eventbus",
		"com.google.common.graph",
		"com.google.common.hash",
		"com.google.common.html",
		"com.google.common.io",
		"com.google.common.math",
		"com.google.common.net",
		"com.google.common.primitives",
		"com.google.common.reflect",
		"com.google.common.util.concurrent",
		"com.google.common.xml",
		"com.google.thirdparty.publicsuffix",
	})
}
