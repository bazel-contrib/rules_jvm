package tsparser

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/parser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/rs/zerolog"
)

func newTestRunner(t *testing.T) (*Runner, string) {
	t.Helper()
	dir := t.TempDir()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	return NewRunner(logger, dir), dir
}

func writeJava(t *testing.T, dir, rel, filename, content string) {
	t.Helper()
	d := filepath.Join(dir, rel)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPackageName(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src/main/java/com/example", "Foo.java", `
package com.example;

public class Foo {}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src/main/java/com/example",
		Files: []string{"Foo.java"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Name.Name != "com.example" {
		t.Errorf("package name = %q, want %q", pkg.Name.Name, "com.example")
	}
}

func TestImports(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "Foo.java", `
package com.example;

import java.util.List;
import java.io.*;
import static org.junit.Assert.assertEquals;
import static com.google.common.collect.ImmutableList.*;

public class Foo {}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"Foo.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Imported classes: java.util.List, org.junit.Assert, com.google.common.collect.ImmutableList
	wantClasses := []string{
		"com.google.common.collect.ImmutableList",
		"java.util.List",
		"org.junit.Assert",
	}
	gotClasses := pkg.ImportedClasses.SortedSlice()
	if len(gotClasses) != len(wantClasses) {
		t.Fatalf("imported classes count = %d, want %d\n  got: %v", len(gotClasses), len(wantClasses), fqns(gotClasses))
	}
	for i, want := range wantClasses {
		if gotClasses[i].FullyQualifiedClassName() != want {
			t.Errorf("imported class[%d] = %q, want %q", i, gotClasses[i].FullyQualifiedClassName(), want)
		}
	}

	// Wildcard package import: java.io
	gotPkgs := pkg.ImportedPackagesWithoutSpecificClasses.SortedSlice()
	if len(gotPkgs) != 1 || gotPkgs[0].Name != "java.io" {
		t.Errorf("imported packages = %v, want [java.io]", gotPkgs)
	}
}

func TestExportedClasses(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "Foo.java", `
package com.example;

public class Foo {}
class Bar {}
`)
	writeJava(t, dir, "src", "Baz.java", `
package com.example;

public interface Baz {}
`)
	writeJava(t, dir, "src", "Status.java", `
package com.example;

public enum Status { OK, ERROR }
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"Foo.java", "Baz.java", "Status.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Only public types: Foo, Baz, Status (not Bar)
	got := pkg.ExportedClasses.SortedSlice()
	want := []string{"com.example.Baz", "com.example.Foo", "com.example.Status"}
	if len(got) != len(want) {
		t.Fatalf("exported classes = %v, want %v", fqns(got), want)
	}
	for i, w := range want {
		if got[i].FullyQualifiedClassName() != w {
			t.Errorf("exported[%d] = %q, want %q", i, got[i].FullyQualifiedClassName(), w)
		}
	}
}

func TestMainMethod(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "App.java", `
package com.example;

public class App {
    public static void main(String[] args) {
        System.out.println("hello");
    }
}
`)
	writeJava(t, dir, "src", "Lib.java", `
package com.example;

public class Lib {
    public void notMain(String[] args) {}
}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"App.java", "Lib.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	got := pkg.Mains.SortedSlice()
	if len(got) != 1 {
		t.Fatalf("mains count = %d, want 1", len(got))
	}
	if got[0].BareOuterClassName() != "App" {
		t.Errorf("main class = %q, want %q", got[0].BareOuterClassName(), "App")
	}
}

func TestMainMethodVarargs(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "App.java", `
package com.example;

public class App {
    public static void main(String... args) {}
}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"App.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pkg.Mains.Len() != 1 {
		t.Fatalf("mains count = %d, want 1", pkg.Mains.Len())
	}
}

func TestAnnotations(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "MyTest.java", `
package com.example;

import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;

@RunWith(SpringRunner.class)
public class MyTest {
    @Autowired
    private String field;

    @Test
    public void testSomething() {}

    @Override
    public String toString() { return ""; }
}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"MyTest.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	meta, ok := pkg.PerClassMetadata["MyTest"]
	if !ok {
		t.Fatal("no PerClassMetadata for MyTest")
	}

	// Class-level: @RunWith
	classAnns := meta.AnnotationClassNames.SortedSlice()
	t.Logf("class annotations: %v", fqns(classAnns))
	if len(classAnns) != 1 || classAnns[0].BareOuterClassName() != "RunWith" {
		t.Errorf("class annotations = %v, want [RunWith]", fqns(classAnns))
	}

	// Method: testSomething has @Test
	testAnns := meta.MethodAnnotationClassNames.Values("testSomething")
	if testAnns == nil || testAnns.Len() != 1 {
		t.Fatalf("testSomething annotations = %v, want 1", testAnns)
	}
	if testAnns.SortedSlice()[0].BareOuterClassName() != "Test" {
		t.Errorf("testSomething annotation = %q, want Test", testAnns.SortedSlice()[0].BareOuterClassName())
	}

	// Method: toString has @Override → java.lang.Override
	overrideAnns := meta.MethodAnnotationClassNames.Values("toString")
	if overrideAnns == nil || overrideAnns.Len() != 1 {
		t.Fatalf("toString annotations = %v, want 1", overrideAnns)
	}
	if overrideAnns.SortedSlice()[0].FullyQualifiedClassName() != "java.lang.Override" {
		t.Errorf("toString annotation = %q, want java.lang.Override", overrideAnns.SortedSlice()[0].FullyQualifiedClassName())
	}

	// Field: field has @Autowired
	fieldAnns := meta.FieldAnnotationClassNames.Values("field")
	if fieldAnns == nil || fieldAnns.Len() != 1 {
		t.Fatalf("field annotations = %v, want 1", fieldAnns)
	}
	if fieldAnns.SortedSlice()[0].BareOuterClassName() != "Autowired" {
		t.Errorf("field annotation = %q, want Autowired", fieldAnns.SortedSlice()[0].BareOuterClassName())
	}
}

func TestTestPackageDetection(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src/test/java/com/example", "FooTest.java", `
package com.example;
public class FooTest {}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src/test/java/com/example",
		Files: []string{"FooTest.java"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !pkg.TestPackage {
		t.Error("expected TestPackage=true for src/test/ path")
	}
}

func TestMultipleFiles(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "A.java", `
package com.example;
import java.util.List;
public class A {}
`)
	writeJava(t, dir, "src", "B.java", `
package com.example;
import java.util.Map;
public class B {}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"A.java", "B.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pkg.ImportedClasses.Len() != 2 {
		t.Errorf("imported classes = %d, want 2", pkg.ImportedClasses.Len())
	}
	if pkg.ExportedClasses.Len() != 2 {
		t.Errorf("exported classes = %d, want 2", pkg.ExportedClasses.Len())
	}
	if pkg.Files.Len() != 2 {
		t.Errorf("files = %d, want 2", pkg.Files.Len())
	}
}

func TestNoPackageDeclaration(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "Script.java", `
public class Script {
    public static void main(String[] args) {}
}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"Script.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pkg.Name.Name != "" {
		t.Errorf("package name = %q, want empty", pkg.Name.Name)
	}
	if pkg.ExportedClasses.Len() != 1 {
		t.Errorf("exported classes = %d, want 1", pkg.ExportedClasses.Len())
	}
	if pkg.Mains.Len() != 1 {
		t.Errorf("mains = %d, want 1", pkg.Mains.Len())
	}
}

func TestMainMethodStrictSignature(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "Good.java", `
package com.example;
public class Good {
    public static void main(String[] args) {}
}
`)
	writeJava(t, dir, "src", "BadExtraArgs.java", `
package com.example;
public class BadExtraArgs {
    public static void main(String[] args, int count) {}
}
`)
	writeJava(t, dir, "src", "BadReturn.java", `
package com.example;
public class BadReturn {
    public static int main(String[] args) { return 0; }
}
`)
	writeJava(t, dir, "src", "BadVisibility.java", `
package com.example;
public class BadVisibility {
    static void main(String[] args) {}
}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel: "src",
		Files: []string{
			"BadExtraArgs.java",
			"BadReturn.java",
			"BadVisibility.java",
			"Good.java",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	got := pkg.Mains.SortedSlice()
	if len(got) != 1 {
		t.Fatalf("mains = %v, want [com.example.Good]", fqns(got))
	}
	if got[0].FullyQualifiedClassName() != "com.example.Good" {
		t.Fatalf("main = %q, want com.example.Good", got[0].FullyQualifiedClassName())
	}
}

func TestStaticNestedImport(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "Foo.java", `
package com.example;

import static com.foo.Outer.Inner.VALUE;
import static com.foo.Util.*;

public class Foo {}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"Foo.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"com.foo.Outer.Inner", "com.foo.Util"}
	got := fqns(pkg.ImportedClasses.SortedSlice())
	if len(got) != len(want) {
		t.Fatalf("imported classes = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("imported classes = %v, want %v", got, want)
		}
	}
}

func TestAnnotationResolutionPrecedence(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "Annotated.java", `
package com.example;

import com.thirdparty.ImportedAnn;

@ImportedAnn
public class Annotated {
	@Deprecated
	void run() {}

	@LocalAnn
	String field;

	@com.scoped.ScopedAnn
	void scoped() {}
}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"Annotated.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := pkg.PerClassMetadata["Annotated"]
	classAnns := fqns(meta.AnnotationClassNames.SortedSlice())
	wantClassAnns := []string{"com.thirdparty.ImportedAnn"}
	if len(classAnns) != len(wantClassAnns) {
		t.Fatalf("class annotations = %v, want %v", classAnns, wantClassAnns)
	}
	for i := range wantClassAnns {
		if classAnns[i] != wantClassAnns[i] {
			t.Fatalf("class annotations = %v, want %v", classAnns, wantClassAnns)
		}
	}

	methodAnns := meta.MethodAnnotationClassNames.Values("run")
	if methodAnns == nil || methodAnns.Len() != 1 {
		t.Fatalf("run method annotations = %v, want [java.lang.Deprecated]", methodAnns)
	}
	if methodAnns.SortedSlice()[0].FullyQualifiedClassName() != "java.lang.Deprecated" {
		t.Fatalf("run method annotation = %q, want java.lang.Deprecated", methodAnns.SortedSlice()[0].FullyQualifiedClassName())
	}

	scopedMethodAnns := meta.MethodAnnotationClassNames.Values("scoped")
	if scopedMethodAnns == nil || scopedMethodAnns.Len() != 1 {
		t.Fatalf("scoped method annotations = %v, want [com.scoped.ScopedAnn]", scopedMethodAnns)
	}
	if scopedMethodAnns.SortedSlice()[0].FullyQualifiedClassName() != "com.scoped.ScopedAnn" {
		t.Fatalf("scoped method annotation = %q, want com.scoped.ScopedAnn", scopedMethodAnns.SortedSlice()[0].FullyQualifiedClassName())
	}

	fieldAnns := meta.FieldAnnotationClassNames.Values("field")
	if fieldAnns == nil || fieldAnns.Len() != 1 {
		t.Fatalf("field annotations = %v, want [com.example.LocalAnn]", fieldAnns)
	}
	if fieldAnns.SortedSlice()[0].FullyQualifiedClassName() != "com.example.LocalAnn" {
		t.Fatalf("field annotation = %q, want com.example.LocalAnn", fieldAnns.SortedSlice()[0].FullyQualifiedClassName())
	}
}

func TestFieldAnnotationsOnMultipleDeclarators(t *testing.T) {
	runner, dir := newTestRunner(t)
	writeJava(t, dir, "src", "Fields.java", `
package com.example;

import javax.inject.Inject;

public class Fields {
	@Inject String a, b;
}
`)

	pkg, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
		Rel:   "src",
		Files: []string{"Fields.java"},
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := pkg.PerClassMetadata["Fields"]
	for _, field := range []string{"a", "b"} {
		fieldAnns := meta.FieldAnnotationClassNames.Values(field)
		if fieldAnns == nil || fieldAnns.Len() != 1 {
			t.Fatalf("field %q annotations = %v, want [javax.inject.Inject]", field, fieldAnns)
		}
		if fieldAnns.SortedSlice()[0].FullyQualifiedClassName() != "javax.inject.Inject" {
			t.Fatalf("field %q annotation = %q, want javax.inject.Inject", field, fieldAnns.SortedSlice()[0].FullyQualifiedClassName())
		}
	}
}

func fqns(cs []types.ClassName) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.FullyQualifiedClassName()
	}
	return out
}
