package parser

import (
	"context"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
)

// ParsePackageRequest describes a directory-relative set of Java sources to parse.
type ParsePackageRequest struct {
	Rel   string
	Files []string
}

// Parser extracts package metadata from Java source files.
type Parser interface {
	ParsePackage(ctx context.Context, in *ParsePackageRequest) (*java.Package, error)
}
