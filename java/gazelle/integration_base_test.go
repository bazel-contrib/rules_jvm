package gazelle

import (
	"errors"
	"fmt"
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
		t.Fatalf("error getting absolute path for workspace")
	}
	c.RepoRoot = absRepoRoot

	for _, lang := range langs {
		cexts = append(cexts, lang)
	}

	return c, langs, cexts
}

// runPerFileTest takes a testing function to test some aspect of Gazelle.
// It configures a test for each subdirectory, and runs the function over them.
// Each test can return a result file, and runPerFileTest will compare it to the file we want.
//
// runPerFileTest assumes that there exists a `testdata` directory under the test's root,
// containing a series of valid Bazel workspaces:
// Example:
//   .
//   ├── test # Test binary
//   └── testdata
//       ├── annotations # A Bazel workspace
//       │   └── WORKSPACE
//       ├── bin # Another Bazel workspace
//       │   └── WORKSPACE
//       └── ...
func runPerFileTest(t *testing.T, testName string, wantFile string, testFunc TestFunc, repositoryHook PrepareRepoFunc) {
	lbl, err := label.Parse(os.Getenv(testTarget))
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
			t.Fatalf("Found non-directory file %s in the root of test data directory %s. Please remove", d.Name(), root)
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

			var prepResult interface{}
			if repositoryHook != nil {
				prepResult, err = repositoryHook(c, langs, cexts)
				if err != nil {
					t.Fatal(err.Error())
				}
			}

			var loads []rule.LoadInfo
			for _, lang := range langs {
				loads = append(loads, lang.Loads()...)
			}
			dir = filepath.Join(wd, dir)
			walk.Walk(c, cexts, []string{dir}, walk.UpdateSubdirsMode, func(dir, rel string, c *config.Config, update bool, oldFile *rule.File, subdirs, regularFiles, genFiles []string) {
				workspaceRelativeDir, err := filepath.Rel(wd, dir)
				if err != nil {
					t.Fatal(err.Error())
				}
				t.Run(rel, func(t *testing.T) {
					f := testFunc(
						langs,
						c,
						dir,
						rel,
						oldFile,
						subdirs,
						regularFiles,
						genFiles,
						workspaceRelativeDir,
						lbl,
						loads,
						prepResult,
					)
					if f != nil {
						got := string(bzl.Format(f.File))
						wantPath := filepath.Join(dir, wantFile)
						workspacePath := filepath.Join(workspace, lbl.Pkg, workspaceRelativeDir, wantFile)
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
							t.Logf("Got: %s", got)
							t.Errorf("%s %q:\n%s", testName, rel, dmp.DiffPrettyText(diffs))
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
								t.Logf("To update golden files, run: bazel run --test_env=%s=true --test_filter='^%s$' %s", writeGoldenFiles, testName, os.Getenv(testTarget))
							}
							t.FailNow()
						}
					}
				})
			})
		})
	}
}

// PrepareRepoFunc is a function executed per repository-under-test.
// It allows you to return any data you want (e.g. a resolution index).
// This data can later be passed in TestFunc's repositoryMetadata argument.
type PrepareRepoFunc func(c *config.Config, langs []language.Language, cexts []config.Configurer) (interface{}, error)

// TestFunc is a function that will be called per-file in a repository whe passed to runPerFileTest.
// It should return a pointer to the resulting file after Gazelle has done its work.
// This file will be later compared to a golden file.
type TestFunc func(
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
	// Value returned from the PrepRepoFunc call before.
	repositoryMetadata interface{},
) *rule.File

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

type testResolver struct{}

func (*testResolver) Resolve(pkg string) (label.Label, error) {
	return label.NoLabel, errors.New("not implemented")
}

type mapResolver map[string]resolve.Resolver

func (mr mapResolver) Resolver(r *rule.Rule, f string) resolve.Resolver {
	return mr[r.Kind()]
}

func InitTestResolversAndExtensions(langs []language.Language) (mapResolver, []interface{}) {
	mrslv := make(mapResolver)
	exts := make([]interface{}, 0, len(langs))
	for _, lang := range langs {
		// TODO There has to be a better way to make this generic.
		if jLang, ok := lang.(*javaLang); ok {
			jLang.mavenResolver = NewTestMavenResolver()
		}

		for kind := range lang.Kinds() {
			mrslv[kind] = lang
		}
		exts = append(exts, lang)
	}
	return mrslv, exts
}

type TestMavenResolver struct {
	data map[string]label.Label
}

func NewTestMavenResolver() *TestMavenResolver {
	return &TestMavenResolver{
		data: map[string]label.Label{
			"com.google.common.primitives": label.New("maven", "", "com_google_guava_guava"),
			"org.junit":                    label.New("maven", "", "junit_junit"),
		},
	}
}

func (r *TestMavenResolver) Resolve(pkg string) (label.Label, error) {
	l, found := r.data[pkg]
	if !found {
		return label.NoLabel, fmt.Errorf("unexpected import: %s", pkg)
	}
	return l, nil
}
