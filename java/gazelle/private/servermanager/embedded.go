package servermanager

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// The embedded javaparser server (and runner script) will be materialized at runtime every time we run gazelle.
// This is not a problem for now, as the javaparser server is tiny.
// However, we may want to get clever and materialize to a more permanent location in the future if the server becomes significant.
//
//go:embed javaparser.jar
var javaparserDeployJar []byte

// GazelleJavaBinEnvVar is an environment variable that clients can use to point to a java installation.
// If this variable or JAVA_HOME are set, the javaparser server will start under that installation of java.
const GazelleJavaBinEnvVar = "GAZELLE_JAVA_JAVAHOME"

func materializeJavaparser(tmpdir string) (string, error) {
	jarPath := filepath.Join(tmpdir, "javaparser.jar")
	err := os.WriteFile(jarPath, javaparserDeployJar, 0644)
	if err != nil {
		return "", err
	}
	return jarPath, nil
}

const unixRunerTemplate = `#!/bin/sh
%s %s -jar %s ${@}
`

const windowsRunnerTemplate = `
@echo off
setlocal

"%s" %s -Xshare:off -jar "%s" %%*

endlocal
`

func createRunner(javaparserLocation string, jvmFlags []string, tmpdir string) (string, error) {
	runnerPath := filepath.Join(tmpdir, "runjavaparser")
	runnerTemplate := unixRunerTemplate

	if runtime.GOOS == "windows" {
		runnerPath += ".bat"
		runnerTemplate = windowsRunnerTemplate
	}

	javaHome := ""
	for _, possibleJavaHome := range []string{GazelleJavaBinEnvVar, "JAVA_HOME"} {
		javaHome = os.Getenv(possibleJavaHome)
		if javaHome != "" {
			break
		}
	}
	if javaHome == "" {
		return "", fmt.Errorf("could not find %s or JAVA_HOME. When running on embedded mode the Gazelle extension for Java requires a java executable. Please set it using %s or disable the extension", GazelleJavaBinEnvVar, GazelleJavaBinEnvVar)
	}

	javaBin := filepath.Join(javaHome, "bin", "java")
	javaparserRunner := fmt.Sprintf(runnerTemplate, javaBin, strings.Join(jvmFlags, " "), javaparserLocation)

	err := os.WriteFile(runnerPath, []byte(javaparserRunner), 0555)
	if err != nil {
		return "", err
	}
	return runnerPath, nil
}

func (m *ServerManager) locateJavaparser(jvmFlags []string) (string, error) {
	javaparserPath, err := materializeJavaparser(m.tmpdir)
	if err != nil {
		return "", fmt.Errorf("failed to materialize java parser: %w", err)
	}

	return createRunner(javaparserPath, jvmFlags, m.tmpdir)
}

func (m *ServerManager) startupFlags(jvmFlags []string) []string {
	return []string{}
}
