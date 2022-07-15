package maven

import (
	"os"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/rs/zerolog"
)

func TestResolver(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Caller().
		Logger().
		Level(zerolog.DebugLevel)

	r, err := NewResolver("testdata/guava_maven_install.json", logger)
	if err != nil {
		t.Fatal(err)
	}

	assertResolves(t, r, "com.google.common.collect", "@maven//:com_google_guava_guava")
	assertResolves(t, r, "javax.annotation", "@maven//:com_google_code_findbugs_jsr305")
	got, err := r.Resolve("unknown.package")
	if err == nil {
		t.Errorf("Want error finding label for unknown.package, got %v", got)
	}
}

func assertResolves(t *testing.T, r Resolver, pkg, wantLabelStr string) {
	got, err := r.Resolve(pkg)
	if err != nil {
		t.Errorf("Error finding label for %v: %v", pkg, err)
	}
	want, _ := label.Parse(wantLabelStr)
	if got != want {
		t.Errorf("Incorrect label for %v; want %v got %v", pkg, want, got)
	}
}
