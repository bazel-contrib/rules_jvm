package servermanager

import (
	"fmt"
	"runtime"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func javaparserLocation() string {
	return "contrib_rules_jvm/java/src/com/github/bazel_contrib/contrib_rules_jvm/javaparser/generators/Main"
}

func (m *ServerManager) locateJavaparser(jvmFlags []string) (string, error) {
	rf, err := runfiles.New()
	if err != nil {
		return "", fmt.Errorf("failed to init new style runfiles: %w", err)
	}

	javaparserPath := javaparserLocation()
	if runtime.GOOS == "windows" {
		javaparserPath += ".exe"
	}
	loc, err := rf.Rlocation(javaparserPath)
	if err != nil {
		return "", fmt.Errorf("failed to call RLocation: %w", err)
	}
	return loc, nil
}

func (m *ServerManager) startupFlags(jvmFlags []string) []string {
	formattedFlags := make([]string, len(jvmFlags))
	for i, flag := range jvmFlags {
		formattedFlags[i] = "--jvm_flag=" + flag
	}
	return formattedFlags
}
