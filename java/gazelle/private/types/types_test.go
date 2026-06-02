package types

import (
	"reflect"
	"sort"
	"testing"
)

func TestParseClassName(t *testing.T) {
	for name, tc := range map[string]struct {
		from    string
		want    *ClassName
		wantErr bool
	}{
		"simple": {
			from: "com.example.Simple",
			want: &ClassName{
				packageName:        NewPackageName("com.example"),
				bareOuterClassName: "Simple",
			},
		},
		"no package": {
			from: "Simple",
			want: &ClassName{
				packageName:        NewPackageName(""),
				bareOuterClassName: "Simple",
			},
		},
		"inner": {
			from: "com.example.Simple.Inner",
			want: &ClassName{
				packageName:        NewPackageName("com.example"),
				bareOuterClassName: "Simple",
				innerClassNames:    []string{"Inner"},
			},
		},
		"nested inner": {
			from: "com.example.Simple.Inner.Nested",
			want: &ClassName{
				packageName:        NewPackageName("com.example"),
				bareOuterClassName: "Simple",
				innerClassNames:    []string{"Inner", "Nested"},
			},
		},
		"anonymous inner": {
			from: "com.example.Simple.",
			want: &ClassName{
				packageName:        NewPackageName("com.example"),
				bareOuterClassName: "Simple",
				innerClassNames:    []string{""},
			},
		},
		"gson class": {
			from: "com.google.gson.Gson",
			want: &ClassName{
				packageName:        NewPackageName("com.google.gson"),
				bareOuterClassName: "Gson",
			},
		},
		"guava strings class": {
			from: "com.google.common.base.Strings",
			want: &ClassName{
				packageName:        NewPackageName("com.google.common.base"),
				bareOuterClassName: "Strings",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := ParseClassName(tc.from)
			if tc.wantErr && err != nil {
				t.Fatal("wanted error, got nil error")
			}
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("wanted no error, got %v", err)
				}
				if !reflect.DeepEqual(*tc.want, *got) {
					t.Fatalf("want %v got %v", tc.want, got)
				}
				if tc.from != got.FullyQualifiedClassName() {
					t.Fatalf("Fully Qualified Class Name: want %s got %s", tc.from, got.FullyQualifiedClassName())
				}
			}
		})
	}
}

func TestClassNameLessMatchesFullyQualifiedStringOrder(t *testing.T) {
	names := []string{
		"A",
		"a.A",
		"a.Z",
		"a.b.A",
		"a.b.C",
		"a.b.C.Inner",
		"a.b.C.Inner.Nested",
		"a.z.A",
		"com.example.Simple",
		"com.example.Simple.Inner",
		"com.google.common.base.Strings",
		"com.google.gson.Gson",
	}

	classes := make([]ClassName, 0, len(names))
	for _, name := range names {
		className, err := ParseClassName(name)
		if err != nil {
			t.Fatalf("ParseClassName(%q): %v", name, err)
		}
		classes = append(classes, *className)
	}

	got := append([]ClassName(nil), classes...)
	sort.Slice(got, func(i, j int) bool {
		return ClassNameLess(got[i], got[j])
	})

	want := append([]ClassName(nil), classes...)
	sort.Slice(want, func(i, j int) bool {
		return want[i].FullyQualifiedClassName() < want[j].FullyQualifiedClassName()
	})

	for i := range want {
		if got[i].FullyQualifiedClassName() != want[i].FullyQualifiedClassName() {
			t.Fatalf("sorted[%d] = %q, want %q", i, got[i].FullyQualifiedClassName(), want[i].FullyQualifiedClassName())
		}
	}
}
