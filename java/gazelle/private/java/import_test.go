package java

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewImport(t *testing.T) {
	tests := []struct {
		in   string
		want Import
	}{
		{
			in: "com.example.api.entities.Principal",
			want: Import{
				Pkg:     "com.example.api.entities",
				Classes: []string{"Principal"},
			},
		},
		{
			in: "com.example.service.cli.utils.Utilities.*",
			want: Import{
				Pkg:     "com.example.service.cli.utils",
				Classes: []string{"Utilities"},
			},
		},
		{
			in: "org.apache.commons.cli.*",
			want: Import{
				Pkg:     "org.apache.commons.cli",
				Classes: []string{},
			},
		},
		{
			in: "com.example.model",
			want: Import{
				Pkg:     "com.example.model",
				Classes: []string{},
			},
		},
		{
			in: "com.example.log.EventLogger.EventType",
			want: Import{
				Pkg:     "com.example.log",
				Classes: []string{"EventLogger", "EventType"},
			},
		},
		{
			in: "autovalue.shaded.com.google$.auto.common.$AnnotationValues$1$1$3",
			want: Import{
				Pkg:     "autovalue.shaded.com.google.auto.common",
				Classes: []string{"AnnotationValues113"},
			},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := NewImport(tt.in)
			if diff := cmp.Diff(*got, tt.want); diff != "" {
				t.Fatalf("NewImport returned diff (-want +got):\n%s", diff)
			}
		})
	}
}
