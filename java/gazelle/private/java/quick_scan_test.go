package java

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/stretchr/testify/require"
)

func TestQuickScanPackage(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected types.PackageName
	}{
		{
			name:     "simple package",
			content:  "package com.example.foo;\n\npublic class Foo {}",
			expected: types.NewPackageName("com.example.foo"),
		},
		{
			name:     "package without semicolon (kotlin style)",
			content:  "package com.example.bar\n\nclass Bar",
			expected: types.NewPackageName("com.example.bar"),
		},
		{
			name:     "package with leading whitespace",
			content:  "  package com.example.baz;\n\nclass Baz {}",
			expected: types.NewPackageName("com.example.baz"),
		},
		{
			name: "package after line comment",
			content: `// This is a comment
package com.example.commented;

class Commented {}`,
			expected: types.NewPackageName("com.example.commented"),
		},
		{
			name: "package after block comment",
			content: `/* This is a
   block comment */
package com.example.block;

class Block {}`,
			expected: types.NewPackageName("com.example.block"),
		},
		{
			name: "package after license header",
			content: `/*
 * Copyright 2025 Example Corp
 * Licensed under Apache 2.0
 */
package com.example.licensed;

public class Licensed {}`,
			expected: types.NewPackageName("com.example.licensed"),
		},
		{
			name: "inline block comment before package",
			content: `/* comment */ package com.example.inline;

class Inline {}`,
			expected: types.NewPackageName("com.example.inline"),
		},
		{
			name:     "no package declaration",
			content:  "public class NoPackage {}",
			expected: types.PackageName{},
		},
		{
			name:     "empty file",
			content:  "",
			expected: types.PackageName{},
		},
		{
			name: "package in comment should be ignored",
			content: `// package com.example.fake;
package com.example.real;

class Real {}`,
			expected: types.NewPackageName("com.example.real"),
		},
		{
			name:     "single segment package",
			content:  "package example;\n\nclass Example {}",
			expected: types.NewPackageName("example"),
		},
		{
			name:     "package with underscore",
			content:  "package com.example_corp.my_pkg;\n\nclass MyClass {}",
			expected: types.NewPackageName("com.example_corp.my_pkg"),
		},
		{
			name:     "package with numbers",
			content:  "package com.example2.v1;\n\nclass V1 {}",
			expected: types.NewPackageName("com.example2.v1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "Test.java")
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			require.NoError(t, err)

			result := QuickScanPackage(filePath)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestQuickScanPackage_NonExistentFile(t *testing.T) {
	result := QuickScanPackage("/nonexistent/path/Foo.java")
	require.Equal(t, types.PackageName{}, result)
}
