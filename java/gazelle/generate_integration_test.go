package gazelle

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/merger"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/bazel-gazelle/testtools"
	"github.com/bazelbuild/bazel-gazelle/walk"
	bzl "github.com/bazelbuild/buildtools/build"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	buildWorkingDirectory = "BUILD_WORKING_DIRECTORY"
	testTarget            = "TEST_TARGET"
	writeGoldenFiles      = "WRITE_GOLDEN_FILES"
)

func shouldWriteGoldenFiles() bool {
	value := os.Getenv(writeGoldenFiles)
	if value == "" {
		return false
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}
	return b
}

type testResolver struct{}

func (*testResolver) Resolve(pkg string) (label.Label, error) {
	return label.NoLabel, errors.New("not implemented")
}

func testConfig(t *testing.T, args ...string) (*config.Config, []language.Language, []config.Configurer) {
	// Add a -repo_root argument if none is present. Without this,
	// config.CommonConfigurer will try to auto-detect a WORKSPACE file,
	// which will fail.
	haveRoot := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-repo_root") {
			haveRoot = true
			break
		}
	}
	if !haveRoot {
		args = append(args, "-repo_root=.")
	}

	cexts := []config.Configurer{
		new(config.CommonConfigurer),
		new(walk.Configurer),
		new(resolve.Configurer),
	}

	l := NewLanguage()
	l.(*javaLang).mavenResolver = &testResolver{}

	langs := []language.Language{
		proto.NewLanguage(),
		l,
	}

	c := testtools.NewTestConfig(t, cexts, langs, args)

	absRepoRoot, err := filepath.Abs(c.RepoRoot)
	if err != nil {
		t.Fatalf("error getting absolute pathRelativeToBazelWorkspaceRoot for workspace")
	}
	c.RepoRoot = absRepoRoot

	for _, lang := range langs {
		cexts = append(cexts, lang)
	}

	return c, langs, cexts
}

func TestGenerateRules(t *testing.T) {
	l, err := label.Parse(os.Getenv(testTarget))
	if err != nil {
		t.Fatal(err.Error())
	}

	workspace := os.Getenv(buildWorkingDirectory)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err.Error())
	}

	const root = "testdata"
	dirEntries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, d := range dirEntries {
		if !d.IsDir() {
			continue
		}
		dir := filepath.Join(root, d.Name())
		t.Run(dir, func(t *testing.T) {
			const inputSuffix = ".in"
			if err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !strings.HasSuffix(path, inputSuffix) {
					return nil
				}

				return os.Rename(path, strings.TrimSuffix(path, inputSuffix))
			}); err != nil {
				t.Fatalf("coud not prepare directory: %s", err)
			}

			c, langs, cexts := testConfig(t, "-repo_root="+dir)

			var loads []rule.LoadInfo
			for _, lang := range langs {
				loads = append(loads, lang.Loads()...)
			}

			walk.Walk(c, cexts, []string{dir}, walk.VisitAllUpdateSubdirsMode, func(dir, rel string, c *config.Config, update bool, oldFile *rule.File, subdirs, regularFiles, genFiles []string) {
				workspaceRelativeDir, err := filepath.Rel(wd, dir)
				if err != nil {
					t.Fatal(err.Error())
				}

				t.Run(rel, func(t *testing.T) {
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
						if name == "BUILD.want" {
							isTest = true
							break
						}
					}
					if !isTest {
						// GenerateRules may have side effects, so we need to run it, even if
						// there's no test.
						return
					}
					f := rule.EmptyFile("test", "")
					for _, r := range gen {
						r.Insert(f)
					}
					convertImportsAttrs(f)
					merger.FixLoads(f, loads)
					f.Sync()
					got := string(bzl.Format(f.File))
					wantPath := filepath.Join(dir, "BUILD.want")
					workspacePath := filepath.Join(workspace, l.Pkg, workspaceRelativeDir, "BUILD.want")
					wantBytes, err := ioutil.ReadFile(wantPath)
					if err != nil {
						t.Fatalf("error reading %s: %v", wantPath, err)
					}
					want := string(wantBytes)
					want = strings.ReplaceAll(want, "\r\n", "\n")
					want = stripTODOs(want)

					if got != want {
						dmp := diffmatchpatch.New()
						diffs := dmp.DiffMain(want, got, true)
						t.Errorf("GenerateRules %q:\n%s", rel, dmp.DiffPrettyText(diffs))
						if shouldWriteGoldenFiles() {
							stat, err := os.Stat(workspacePath)
							if err != nil {
								t.Fatalf("error opening golden file: %v", err)
							}
							if err := os.WriteFile(workspacePath, []byte(got), stat.Mode()); err != nil {
								t.Fatalf("error writing golden file: %s", err)
							}
							t.Logf("Updated the golden file: %s", workspacePath)
						} else {
							t.Logf("To update golden files, run: bazel run --test_env=%s=true --test_filter='^TestGenerateRules$' %s", writeGoldenFiles, os.Getenv(testTarget))
						}
						t.FailNow()
					}
				})
			})
		})
	}
}

func stripTODOs(s string) string {
	inputLines := strings.Split(s, "\n")
	outputLines := make([]string, 0, len(inputLines))
	for _, line := range inputLines {
		split := strings.SplitN(line, "# TODO", 2)
		if len(split) == 1 {
			outputLines = append(outputLines, line)
		} else if len(split) == 2 {
			before := strings.TrimRight(split[0], " ")
			if len(before) > 0 {
				outputLines = append(outputLines, before)
			}
		}
	}
	return strings.Join(outputLines, "\n")
}

// convertImportsAttrs copies private attributes to regular attributes, which
// will later be written out to build files. This allows tests to check the
// values of private attributes with simple string comparison.
func convertImportsAttrs(f *rule.File) {
	copyAttributes := []string{
		config.GazelleImportsKey,
		packagesKey,
	}

	for _, r := range f.Rules {
		for _, a := range copyAttributes {
			v := r.PrivateAttr(a)
			if v != nil {
				r.SetAttr(a, v)
			}
		}
	}
}
