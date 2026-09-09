package gazelle_test

import (
	"testing"

	javacontrib "github.com/bazel-contrib/rules_jvm/java/gazelle"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/bazel-gazelle/testtools"
	bzl "github.com/bazelbuild/buildtools/build"
)

func TestResolveToJavaExport(t *testing.T) {
	tests := map[string]struct {
		boundaryKind string
		kindMap      map[string]config.MappedKind
		srcs         []string
		computedSrcs bool
		disabled     bool
		sameBoundary bool
		want         label.Label
	}{
		"java export": {
			boundaryKind: "java_export",
			want:         label.New("", "", "boundary"),
		},
		"transitively mapped java export": {
			boundaryKind: "custom_java_export",
			kindMap: map[string]config.MappedKind{
				"java_export":              {FromKind: "java_export", KindName: "intermediate_java_export"},
				"intermediate_java_export": {FromKind: "intermediate_java_export", KindName: "custom_java_export"},
			},
			want: label.New("", "", "boundary"),
		},
		"java export before map application": {
			boundaryKind: "java_export",
			kindMap: map[string]config.MappedKind{
				"java_export":              {FromKind: "java_export", KindName: "intermediate_java_export"},
				"intermediate_java_export": {FromKind: "intermediate_java_export", KindName: "custom_java_export"},
			},
			want: label.New("", "", "boundary"),
		},
		"self-mapped java export": {
			boundaryKind: "java_export",
			kindMap: map[string]config.MappedKind{
				"java_export": {FromKind: "java_export", KindName: "java_export"},
			},
			want: label.New("", "", "boundary"),
		},
		"source-less kotlin export": {
			boundaryKind: "kt_jvm_export",
			want:         label.New("", "", "boundary"),
		},
		"transitively mapped source-less kotlin export": {
			boundaryKind: "custom_kt_jvm_export",
			kindMap: map[string]config.MappedKind{
				"kt_jvm_export":              {FromKind: "kt_jvm_export", KindName: "intermediate_kt_jvm_export"},
				"intermediate_kt_jvm_export": {FromKind: "intermediate_kt_jvm_export", KindName: "custom_kt_jvm_export"},
			},
			want: label.New("", "", "boundary"),
		},
		"source-less kotlin export before map application": {
			boundaryKind: "kt_jvm_export",
			kindMap: map[string]config.MappedKind{
				"kt_jvm_export":              {FromKind: "kt_jvm_export", KindName: "intermediate_kt_jvm_export"},
				"intermediate_kt_jvm_export": {FromKind: "intermediate_kt_jvm_export", KindName: "custom_kt_jvm_export"},
			},
			want: label.New("", "", "boundary"),
		},
		"source-bearing kotlin export": {
			boundaryKind: "kt_jvm_export",
			srcs:         []string{"Boundary.kt"},
			want:         label.New("", "", "dep_java_library"),
		},
		"computed kotlin sources": {
			boundaryKind: "kt_jvm_export",
			computedSrcs: true,
			want:         label.New("", "", "dep_java_library"),
		},
		"no export": {
			want: label.New("", "", "dep_java_library"),
		},
		"resolution disabled": {
			boundaryKind: "java_export",
			disabled:     true,
			want:         label.New("", "", "dep_java_library"),
		},
		"same export": {
			boundaryKind: "java_export",
			sameBoundary: true,
			want:         label.New("", "", "dep_java_library"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			javaLanguage := javacontrib.NewLanguage()
			c := testtools.NewTestConfig(t, nil, []language.Language{javaLanguage}, nil)
			c.KindMap = tt.kindMap
			if tt.disabled {
				c.Exts["java"].(javaconfig.Configs)[""].SetResolveToJavaExports(false)
			}
			queryConfig := c.Clone()

			f := rule.EmptyFile("BUILD.bazel", "")
			if tt.boundaryKind != "" {
				boundary := rule.NewRule(tt.boundaryKind, "boundary")
				deps := []string{":dep_java_library"}
				if tt.sameBoundary {
					deps = append(deps, ":consumer_java_library")
				}
				boundary.SetAttr("deps", deps)
				if tt.computedSrcs {
					boundary.SetAttr("srcs", &bzl.CallExpr{X: &bzl.Ident{Name: "glob"}})
				} else {
					boundary.SetAttr("srcs", tt.srcs)
				}
				boundary.Insert(f)
			}
			javaLanguage.Fix(c, f)

			protoRules := []*rule.Rule{
				protoLibrary("dep_proto", "com.example.dep"),
				protoLibrary("consumer_proto", "com.example.consumer"),
			}
			javaLanguage.GenerateRules(language.GenerateArgs{
				Config:   c,
				File:     f,
				OtherGen: protoRules,
			})
			javaLanguage.(language.LifecycleManager).DoneGeneratingRules()

			candidate := resolve.FindResult{Label: label.New("", "", "dep_java_library")}
			got := javacontrib.ResolveToJavaExport(
				queryConfig,
				[]resolve.FindResult{candidate},
				label.New("", "", "consumer_java_library"),
			)
			if len(got) != 1 || got[0].Label != tt.want {
				t.Fatalf("ResolveToJavaExport() = %v, want %s", got, tt.want)
			}
		})
	}
}

func protoLibrary(name, javaPackage string) *rule.Rule {
	r := rule.NewRule("proto_library", name)
	r.SetPrivateAttr(proto.PackageKey, proto.Package{
		Files:   map[string]proto.FileInfo{},
		Options: map[string]string{"java_package": javaPackage},
	})
	return r
}
