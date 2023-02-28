package main

import (
	"io"
	"log"
	"os"
	"text/template"

	"go.starlark.net/starlark"
)

const tpl = `package javaconfig

var defaultTestFileSuffixes = []string{
{{- range .TestFileSuffixes }}
	"{{.}}",
{{- end }}
}
`

type tplData struct {
	TestFileSuffixes []string
}

func main() {
	var inputPath, outputPath string
	switch len(os.Args) {
	case 2:
		inputPath = os.Args[1]
		outputPath = ""

	case 3:
		inputPath = os.Args[1]
		outputPath = os.Args[2]

	default:
		log.Fatalf("Want 2 args, got: %d: %v", len(os.Args)-1, os.Args[1:])
	}

	src, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to read %v: %v", inputPath, err)
	}

	globals, err := starlark.ExecFile(new(starlark.Thread), inputPath, src, nil)
	if err != nil {
		log.Fatalf("starlark error: %s", err)
	}

	var data tplData
	iter := globals["DEFAULT_TEST_SUFFIXES"].(*starlark.List).Iterate()
	defer iter.Done()
	var x starlark.Value
	for iter.Next(&x) {
		data.TestFileSuffixes = append(data.TestFileSuffixes, x.(starlark.String).GoString())
	}

	if len(data.TestFileSuffixes) == 0 {
		log.Fatalf("Didn't find any DEFAULT_TEST_SUFFIXES values")
	}

	t, err := template.New("").Parse(tpl)
	if err != nil {
		log.Fatalf("template parse error: %s", err)
	}

	var out io.Writer
	if outputPath == "" {
		out = os.Stdout
	} else {
		f, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("file create error: %s", err)
		}
		defer f.Close()
		out = f
	}

	if err := t.Execute(out, data); err != nil {
		log.Fatalf("template parse error: %s", err)
	}
}
