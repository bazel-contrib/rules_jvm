package types

import (
	"reflect"
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
