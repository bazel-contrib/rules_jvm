package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/cmd/parsejars/manifest"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/bazel"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/logconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/rs/zerolog"
)

func main() {
	var relativeMavenInstallFile, relativeOutput, repoRoot string

	flag.StringVar(&relativeMavenInstallFile, "maven-install", "maven_install.json", "maven install file (relative to repo root)")
	flag.StringVar(&repoRoot, "repo-root", ".", "repo root")
	flag.StringVar(&relativeOutput, "output", "maven_manifest.json", "output (relative to repo root)")
	flag.Parse()

	goLevel, _ := logconfig.LogLevel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(goLevel)
	logger.Print("creating java language")

	mavenInstallFile := filepath.Join(repoRoot, relativeMavenInstallFile)
	output := filepath.Join(repoRoot, relativeOutput)

	outputBase, err := bazel.OutputBase(repoRoot)
	if err != nil {
		log.Fatalln("error getting output_base")
	}

	resolver, err := maven.NewResolver(outputBase, mavenInstallFile, logger)
	if err != nil {
		log.Fatalln("error creating Maven resolver")
	}

	rd := resolver.(maven.Manifester)
	m := rd.DumpManifest()

	f, err := os.Create(output)
	if err != nil {
		log.Fatalf("file error: %s", err)
	}
	defer f.Close()

	if err := manifest.NewFile(m).Encode(f, mavenInstallFile); err != nil {
		log.Fatalf("error: %s", err)
	}
}
