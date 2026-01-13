package java

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

// packageRegex matches Java/Kotlin package declarations.
// Handles: package com.example.foo;
//
//	package com.example.foo
var packageRegex = regexp.MustCompile(`^\s*package\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)*)\s*;?\s*$`)

// QuickScanPackage performs a quick scan of a Java or Kotlin source file to extract
// the package declaration. It only reads the first portion of the file, looking for
// the package statement before any class/interface/enum declarations.
//
// Returns the package name if found, or an empty PackageName if not found or on error.
func QuickScanPackage(filePath string) types.PackageName {
	file, err := os.Open(filePath)
	if err != nil {
		return types.PackageName{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inBlockComment := false

	for scanner.Scan() {
		line := scanner.Text()

		// Handle block comments
		if inBlockComment {
			if idx := strings.Index(line, "*/"); idx != -1 {
				inBlockComment = false
				line = line[idx+2:]
			} else {
				continue
			}
		}

		// Remove block comment starts
		for {
			startIdx := strings.Index(line, "/*")
			if startIdx == -1 {
				break
			}
			endIdx := strings.Index(line[startIdx+2:], "*/")
			if endIdx == -1 {
				inBlockComment = true
				line = line[:startIdx]
				break
			}
			line = line[:startIdx] + line[startIdx+2+endIdx+2:]
		}

		// Remove line comments
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for package declaration
		if matches := packageRegex.FindStringSubmatch(line); matches != nil {
			return types.NewPackageName(matches[1])
		}

		// If we hit an import, class, interface, enum, or annotation declaration,
		// stop scanning - we've passed where the package would be declared
		if strings.HasPrefix(line, "import ") ||
			strings.HasPrefix(line, "class ") ||
			strings.HasPrefix(line, "public ") ||
			strings.HasPrefix(line, "private ") ||
			strings.HasPrefix(line, "protected ") ||
			strings.HasPrefix(line, "abstract ") ||
			strings.HasPrefix(line, "final ") ||
			strings.HasPrefix(line, "interface ") ||
			strings.HasPrefix(line, "enum ") ||
			strings.HasPrefix(line, "@") {
			break
		}
	}

	return types.PackageName{}
}
