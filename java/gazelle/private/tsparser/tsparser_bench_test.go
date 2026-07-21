package tsparser

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/parser"
	"github.com/rs/zerolog"
)

type javaCorpusFile struct {
	path   string
	rel    string
	name   string
	source []byte
}

type javaCorpusPackage struct {
	rel   string
	files []string
}

func benchmarkCorpusEnv(primary, fallback string) string {
	if value := os.Getenv(primary); value != "" {
		return value
	}
	return os.Getenv(fallback)
}

func loadJavaBenchmarkCorpus(b *testing.B) (string, []javaCorpusFile) {
	b.Helper()

	root := benchmarkCorpusEnv("RULES_JVM_JAVA_CORPUS", "GOT_JAVA_CORPUS")
	if root == "" {
		b.Skip("set RULES_JVM_JAVA_CORPUS to a Java corpus root")
	}

	var files []javaCorpusFile
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".java" {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, javaCorpusFile{
			path:   path,
			rel:    rel,
			name:   filepath.Base(path),
			source: source,
		})
		return nil
	}); err != nil {
		b.Fatal(err)
	}
	if len(files) == 0 {
		b.Fatalf("no Java files under %s", root)
	}

	switch benchmarkCorpusEnv("RULES_JVM_JAVA_CORPUS_ORDER", "GOT_JAVA_CORPUS_ORDER") {
	case "random":
		seed := int64(1)
		if raw := benchmarkCorpusEnv("RULES_JVM_JAVA_CORPUS_RANDOM_SEED", "GOT_JAVA_CORPUS_RANDOM_SEED"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				b.Fatal(err)
			}
			seed = parsed
		}
		r := rand.New(rand.NewSource(seed))
		r.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] })
	default:
		sort.Slice(files, func(i, j int) bool {
			if len(files[i].source) == len(files[j].source) {
				return files[i].path < files[j].path
			}
			return len(files[i].source) > len(files[j].source)
		})
	}

	maxFiles := 10
	if raw := benchmarkCorpusEnv("RULES_JVM_JAVA_CORPUS_MAX_FILES", "GOT_JAVA_CORPUS_MAX_FILES"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			b.Fatal(err)
		}
		maxFiles = parsed
	}
	if maxFiles > 0 && len(files) > maxFiles {
		files = files[:maxFiles]
	}
	return root, files
}

func groupJavaCorpusPackages(files []javaCorpusFile) []javaCorpusPackage {
	byRel := make(map[string][]string)
	for _, file := range files {
		rel := filepath.Dir(file.rel)
		if rel == "." {
			rel = ""
		}
		byRel[rel] = append(byRel[rel], file.name)
	}

	packages := make([]javaCorpusPackage, 0, len(byRel))
	for rel, names := range byRel {
		sort.Strings(names)
		packages = append(packages, javaCorpusPackage{rel: rel, files: names})
	}
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].rel < packages[j].rel
	})
	return packages
}

func javaCorpusBytes(files []javaCorpusFile) int64 {
	var total int64
	for _, file := range files {
		total += int64(len(file.source))
	}
	return total
}

func BenchmarkJavaCorpusParsePackage(b *testing.B) {
	root, files := loadJavaBenchmarkCorpus(b)
	packages := groupJavaCorpusPackages(files)
	b.ReportAllocs()
	b.SetBytes(javaCorpusBytes(files))
	b.ResetTimer()

	runner := NewRunner(zerolog.Nop(), root)
	var errors int64
	for i := 0; i < b.N; i++ {
		for _, pkg := range packages {
			_, err := runner.ParsePackage(context.Background(), &parser.ParsePackageRequest{
				Rel:   pkg.rel,
				Files: pkg.files,
			})
			if err != nil {
				errors++
			}
		}
	}
	b.ReportMetric(float64(errors)/float64(b.N), "errors/op")
}

func BenchmarkJavaCorpusParseOnlyFullTree(b *testing.B) {
	_, files := loadJavaBenchmarkCorpus(b)
	runner := NewRunner(zerolog.Nop(), "")
	b.ReportAllocs()
	b.SetBytes(javaCorpusBytes(files))
	b.ResetTimer()

	var errors int64
	for i := 0; i < b.N; i++ {
		for _, file := range files {
			tree, err := runner.parseJavaContent(file.source)
			if err != nil || tree == nil {
				errors++
				continue
			}
			tree.Release()
		}
	}
	b.ReportMetric(float64(errors)/float64(b.N), "errors/op")
}

func BenchmarkJavaCorpusParseOnlyNoTree(b *testing.B) {
	_, files := loadJavaBenchmarkCorpus(b)
	b.ReportAllocs()
	b.SetBytes(javaCorpusBytes(files))
	b.ResetTimer()

	var errors int64
	for i := 0; i < b.N; i++ {
		for _, file := range files {
			tree, err := tsJavaParserPool.ParseNoTreeBenchmarkOnly(file.source)
			if err != nil || tree == nil {
				errors++
				continue
			}
			tree.Release()
		}
	}
	b.ReportMetric(float64(errors)/float64(b.N), "errors/op")
}
