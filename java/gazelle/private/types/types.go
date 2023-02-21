package types

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
)

type PackageName struct {
	Name string
}

func NewPackageName(name string) PackageName {
	return PackageName{
		Name: name,
	}
}

func PackageNameLess(l, r PackageName) bool {
	return l.Name < r.Name
}

func PackageNamesHasPrefix(whole, prefix PackageName) bool {
	return whole.Name == prefix.Name || strings.HasPrefix(whole.Name, prefix.Name+".")
}

type ClassName struct {
	packageName PackageName
	/// bareOuterClassName is the top-most class. i.e. for `com.example.OuterClass.InnerClass.EvenMoreInnerClass`, this will be `OuterClass`.
	bareOuterClassName string
	/// innerClassNames contains all of the class-names nested inside bareOuterClassName, but excluding bareOuterClassName.
	/// i.e. for `com.example.OuterClass.InnerClass.EvenMoreInnerClass`, this will be ["InnerClass", "EvenMoreInnerClass"].
	innerClassNames []string
}

func (c *ClassName) PackageName() PackageName {
	return c.packageName
}

func (c *ClassName) BareOuterClassName() string {
	return c.bareOuterClassName
}

func (c *ClassName) FullyQualifiedClassName() string {
	var parts []string
	if c.packageName.Name != "" {
		parts = append(parts, strings.Split(c.packageName.Name, ".")...)
	}
	parts = append(parts, c.bareOuterClassName)
	parts = append(parts, c.innerClassNames...)
	return strings.Join(parts, ".")
}

func NewClassName(packageName PackageName, bareOuterClassName string) ClassName {
	return ClassName{
		packageName:        packageName,
		bareOuterClassName: bareOuterClassName,
	}
}

func ParseClassName(fullyQualified string) (*ClassName, error) {
	parts := strings.Split(fullyQualified, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("failed to parse class name: empty")
	}

	indexOfOuterClassName := len(parts) - 1
	for i := len(parts) - 1; i >= 0; i-- {
		if unicode.IsUpper([]rune(parts[i])[0]) {
			indexOfOuterClassName = i
		} else {
			break
		}
	}

	packageName := NewPackageName(strings.Join(parts[:indexOfOuterClassName], "."))

	var innerClassNames []string
	if indexOfOuterClassName < (len(parts) - 1) {
		innerClassNames = parts[indexOfOuterClassName+1:]
	}

	return &ClassName{
		packageName:        packageName,
		bareOuterClassName: parts[indexOfOuterClassName],
		innerClassNames:    innerClassNames,
	}, nil
}

func ClassNameLess(l, r ClassName) bool {
	return l.FullyQualifiedClassName() < r.FullyQualifiedClassName()
}

type ResolveInput struct {
	PackageNames         *sorted_set.SortedSet[PackageName]
	ImportedPackageNames *sorted_set.SortedSet[PackageName]
}

type ResolvableJavaPackage struct {
	packageName PackageName
	isTestOnly  bool
}

func NewResolvableJavaPackage(packageName PackageName, isTestOnly bool) ResolvableJavaPackage {
	return ResolvableJavaPackage{
		packageName: packageName,
		isTestOnly:  isTestOnly,
	}
}

func (r *ResolvableJavaPackage) PackageName() PackageName {
	return r.packageName
}

func (r *ResolvableJavaPackage) String() string {
	s := r.packageName.Name
	if r.isTestOnly {
		s += "!testonly"
	}
	return s
}

func ParseResolvableJavaPackage(s string) (*ResolvableJavaPackage, error) {
	parts := strings.Split(s, "!")
	if len(parts) > 2 {
		return nil, fmt.Errorf("want 1 or 2 parts separated by !, got: %q", s)
	}
	packageName := NewPackageName(parts[0])
	isTestOnly := false
	for _, part := range parts[1:] {
		if part == "testonly" {
			isTestOnly = true
		} else {
			return nil, fmt.Errorf("saw unrecognized tag %s", part)
		}
	}
	return &ResolvableJavaPackage{
		packageName: packageName,
		isTestOnly:  isTestOnly,
	}, nil
}
