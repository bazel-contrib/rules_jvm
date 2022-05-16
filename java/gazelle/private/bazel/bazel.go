package bazel

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

var (
	FindBinary   = bazel.FindBinary
	ListRunfiles = bazel.ListRunfiles
)

var nonWordRe = regexp.MustCompile(`\W+`)

func CleanupLabel(in string) string {
	return nonWordRe.ReplaceAllString(in, "_")
}

// bazelCmd copied from bazelbuild/rules_go
// https://github.com/bazelbuild/rules_go/blob/b6f1c12b45f0aa85a4221c4f200f83de21e2ac4d/go/tools/bazel_testing/bazel_testing.go#L162-L180
func bazelCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("bazel")
	if testTempDir, ok := os.LookupEnv(bazel.TEST_TMPDIR); ok {
		cmd.Args = append(cmd.Args, "--output_user_root="+testTempDir)
	}
	cmd.Args = append(cmd.Args, args...)
	for _, e := range os.Environ() {
		// Filter environment variables set by the bazel test wrapper script.
		// These confuse recursive invocations of Bazel.
		if strings.HasPrefix(e, "TEST_") || strings.HasPrefix(e, "RUNFILES_") {
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}
	return cmd
}

// Returns the output from a bazel info command
func bazelInfo(repoRoot string, information string) (string, error) {
	var cmdOut, cmdErr bytes.Buffer
	cmd := bazelCmd("info", information)
	cmd.Dir = repoRoot
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("bazel info error: %s: %s", err, cmdErr.String())
	}

	return strings.TrimSpace(cmdOut.String()), nil
}

func OutputBase(repoRoot string) (string, error) {
	return bazelInfo(repoRoot, "output_base")
}
