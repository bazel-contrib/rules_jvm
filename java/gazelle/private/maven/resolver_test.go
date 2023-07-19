package maven

import (
	"os"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/rs/zerolog"
)

func TestResolver(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Caller().
		Logger().
		Level(zerolog.DebugLevel)

	m := make(map[string]struct{})
	m["@maven//:com_google_j2objc_j2objc_annotations"] = struct{}{}
	r, err := NewResolver("testdata/guava_maven_install.json", logger)
	if err != nil {
		t.Fatal(err)
	}

	assertResolves(t, r, m, "com.google.common.collect", "@maven//:com_google_guava_guava")
	assertResolves(t, r, m, "javax.annotation", "@maven//:com_google_code_findbugs_jsr305")
	got, err := r.Resolve(types.NewPackageName("unknown.package"), m, "maven")
	if err == nil {
		t.Errorf("Want error finding label for unknown.package, got %v", got)
	}
	got, err = r.Resolve(types.NewPackageName("com.google.j2objc.annotations"), m, "maven")
	if err == nil {
		t.Errorf("Want error finding label for excluded artifact, got %v", got)
	}

}

func assertResolves(t *testing.T, r Resolver, excludePackages map[string]struct{}, pkg, wantLabelStr string) {
	got, err := r.Resolve(types.NewPackageName(pkg), excludePackages, "maven")
	if err != nil {
		t.Errorf("Error finding label for %v: %v", pkg, err)
	}
	want, _ := label.Parse(wantLabelStr)
	if got != want {
		t.Errorf("Incorrect label for %v; want %v got %v", pkg, want, got)
	}
}
