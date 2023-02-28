package main

import (
	"log"
	"os"
	"reflect"

	bzl "github.com/bazelbuild/buildtools/build"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Want 2 args, got: %d: %v", len(os.Args)-1, os.Args[1:])
	}
	inputPath := os.Args[1]
	outputPath := os.Args[2]

	contents, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to read %v: %v", inputPath, err)
	}
	file, err := bzl.ParseBzl(inputPath, contents)
	if err != nil {
		log.Fatalf("Failed to parse %v: %v", inputPath, err)
	}

	var defaultTestSuffixes []string

	for _, expr := range file.Stmt {
		assign, ok := expr.(*bzl.AssignExpr)
		if !ok {
			continue
		}
		ident, ok := assign.LHS.(*bzl.Ident)
		if !ok {
			continue
		}
		if ident.Name != "DEFAULT_TEST_SUFFIXES" {
			continue
		}
		if len(defaultTestSuffixes) > 0 {
			log.Fatal("Expected only one DEFAULT_TEST_SUFFIXES assignment but saw multiple")
		}
		rhs, ok := assign.RHS.(*bzl.ListExpr)
		if !ok {
			log.Fatalf("DEFAULT_TEST_SUFFIXES must be a list but got %v", reflect.TypeOf(assign.RHS).Name())
		}
		for _, expr := range rhs.List {
			lit, ok := expr.(*bzl.StringExpr)
			if !ok {
				log.Fatalf("DEFAULT_TEST_SUFFIXES values must be string literals but got %v", reflect.TypeOf(expr))
			}
			defaultTestSuffixes = append(defaultTestSuffixes, lit.Value)
		}
	}
	if len(defaultTestSuffixes) == 0 {
		log.Fatalf("Didn't find any DEFAULT_TEST_SUFFIXES values")
	}
	output := `package javaconfig

var defaultTestFileSuffixes = []string{
`

	for _, dts := range defaultTestSuffixes {
		output += "\t\"" + dts + "\",\n"
	}

	output += "}\n"
	if err := os.WriteFile(outputPath, []byte(output), 0o644); err != nil {
		log.Fatal(err)
	}
}
