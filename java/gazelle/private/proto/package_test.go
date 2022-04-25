package proto

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseFile(t *testing.T) {
	tests := map[string]struct {
		want File
	}{
		"simple": {
			want: File{
				PackageName: "com.example.book",
				Options: map[string]string{
					"go_package":          "example.com/books;books",
					"java_multiple_files": "true",
					"java_package":        "com.example.book",
				},
				Enums:    []string{"BookType"},
				Messages: []string{"ReadBookRequest", "ReadBookResponse"},
				Services: []string{"Books"},
			},
		},
		"multi-line-option": {
			want: File{
				PackageName: "com.example.book",
				Options: map[string]string{
					"go_package":          "example.com/books;books",
					"java_multiple_files": "true",
					"java_package":        "com.example.book",
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			filename := filepath.Join("testdata", name+".proto")
			got, err := ParseFile(filename)
			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}
			if diff := cmp.Diff(*got, tt.want); diff != "" {
				t.Fatalf("ParseFile returned diff (-want +got):\n%s", diff)
			}
		})
	}
}
