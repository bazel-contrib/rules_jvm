package gazelle

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/bazel-gazelle/walk"
)

const resolveWantFile = "BUILD.want.res"

type PrepResult struct {
	resolver mapResolver
	index    *resolve.RuleIndex
}

func PrepareRepoForResolve(c *config.Config, langs []language.Language, cexts []config.Configurer) (interface{}, error) {
	root := c.RepoRoot
	mrslv, exts := InitTestResolversAndExtensions(langs)

	// We need to do our own traversal, because we need a full index to resolve a single rule.
	// We populate the index here.
	ix := resolve.NewRuleIndex(mrslv.Resolver, exts...)
	// We need to walk from the absolute root, because `Walk` falls back to c.RepoRoot as the root if it can't find the dirs we want to walk.
	absRoot := filepath.Join(c.RepoRoot, root)
	var walkErr error = nil
	walk.Walk(c, cexts, []string{absRoot}, walk.UpdateSubdirsMode, func(dir, _ string, c *config.Config, _ bool, _ *rule.File, _, _, _ []string) {
		// walk.WalkFunc does not have a return, so we can't return an error.
		if walkErr != nil {
			return
		}
		buildPath := filepath.Join(dir, generateWantFile)
		pkg, err := filepath.Rel(c.RepoRoot, dir)
		if err != nil {
			walkErr = err
		}
		f, err := rule.LoadFile(buildPath, pkg)
		if f == nil {
			walkErr = fmt.Errorf("Tried to load non-existing file %s, from directory dir %s", buildPath, dir)
		}
		if err != nil {
			walkErr = err
		}
		for _, r := range f.Rules {
			// Here, `f` points to the generated file.
			ix.AddRule(c, r, f)
		}
	})
	if walkErr != nil {
		return nil, walkErr
	}
	ix.Finish()
	return PrepResult{resolver: mrslv, index: ix}, nil
}

func TestResolveRules(t *testing.T) {
	testName := "TestResolveRules"

	// For each file we're testing, load the output of `Generate`, and resolve it with the index.
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
		extraRepoInfo interface{},
	) *rule.File {
		prep := extraRepoInfo.(PrepResult)
		ix := prep.index
		mrslv := prep.resolver
		buildPath := filepath.Join(dir, generateWantFile)
		f, err := rule.LoadFile(buildPath, dir)
		if err != nil {
			return nil
		}
		imports := make([]interface{}, 0, len(f.Rules))
		for _, r := range f.Rules {
			value := r.AttrStrings(config.GazelleImportsKey)
			r.DelAttr(config.GazelleImportsKey)
			r.DelAttr(packagesKey)
			imports = append(imports, value)
			mrslv.Resolver(r, "").Resolve(c, ix, nil, r, value, label.New("", workspaceRelativeDir, r.Name()))
		}
		f.Sync()
		return f
	}

	runPerFileTest(t, testName, resolveWantFile, test, PrepareRepoForResolve)
}
