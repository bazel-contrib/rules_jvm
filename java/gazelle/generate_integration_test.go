package gazelle

import (
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/merger"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

const generateWantFile = "BUILD.want.gen"

func TestGenerateRules(t *testing.T) {
	testName := "TestGenerateRules"
	var test TestFunc = func(
		langs []language.Language,
		c *config.Config,
		dir string,
		rel string,
		oldFile *rule.File,
		subdirs,
		regularFiles,
		genFiles []string,
		workspaceRelativeDir string,
		lbl label.Label,
		loads []rule.LoadInfo,
		// This would hold whatever our per-repo initialization returned.
		// Since Generate doesn't need any per-repo initialization, we ignore this result.
		_ interface{},
	) *rule.File {
		var gen []*rule.Rule
		var empty []*rule.Rule
		var imports []interface{}
		for _, lang := range langs {
			res := lang.GenerateRules(language.GenerateArgs{
				Config:       c,
				Dir:          dir,
				Rel:          rel,
				File:         oldFile,
				Subdirs:      subdirs,
				RegularFiles: regularFiles,
				GenFiles:     genFiles,
				OtherEmpty:   empty,
				OtherGen:     gen,
			})
			gen = append(gen, res.Gen...)
			empty = append(empty, res.Empty...)
			imports = append(imports, res.Imports...)

			for i := 0; i < len(gen); i++ {
				gen[i].SetPrivateAttr(config.GazelleImportsKey, imports[i])
			}
		}
		isTest := false
		for _, name := range regularFiles {
			if name == generateWantFile {
				isTest = true
				break
			}
		}
		if !isTest {
			// GenerateRules may have side effects, so we need to run it, even if
			// there's no test.
			return nil
		}
		f := rule.EmptyFile("test", "")
		for _, r := range gen {
			r.Insert(f)
		}
		convertImportsAttrs(f)
		merger.FixLoads(f, loads)
		f.Sync()
		return f
	}
	runPerFileTest(t, testName, generateWantFile, test, nil)
}
