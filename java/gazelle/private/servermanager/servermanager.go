package servermanager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"google.golang.org/grpc"
)

const javaparser = "java/src/com/github/bazel_contrib/contrib_rules_jvm/javaparser/generators/Main"

type ServerManager struct {
	workspace    string
	javaLogLevel string

	mu   sync.Mutex
	conn *grpc.ClientConn
}

func New(workspace, javaLogLevel string) *ServerManager {
	return &ServerManager{
		workspace:    workspace,
		javaLogLevel: javaLogLevel,
	}
}

func (m *ServerManager) Connect() (*grpc.ClientConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil {
		return m.conn, nil
	}

	logLevelFlag := fmt.Sprintf("-Dorg.slf4j.simpleLogger.defaultLogLevel=%s", m.javaLogLevel)

	javaParserPath, err := locateJavaparser()
	if err != nil {
		return nil, fmt.Errorf("failed to find javaparser in runfiles: %w", err)
	}

	dir, err := os.MkdirTemp(os.Getenv("TMPDIR"), "gazelle-javaparser")
	if err != nil {
		return nil, fmt.Errorf("failed to create tmpdir to start javaparser server: %w", err)
	}
	portFilePath := filepath.Join(dir, "port")

	cmd := exec.Command(
		javaParserPath,
		"--jvm_flag="+logLevelFlag,
		"--server-port-file-path", portFilePath,
		"--workspace", m.workspace,
		"--idle-timeout", "30")
	// Send JavaParser stdout to stderr for two reasons:
	//  1. We don't want to pollute our own stdout
	//  2. Java does some output buffer sniffing where it will block its own progress until the
	//     stdout buffer is read from, whereas stderr is unbuffered so doesn't hit this issue.
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start javaparser sever: %w", err)
	}

	port, err := readPort(portFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read port from javaparser server - maybe it crashed: %w", err)
	}

	addr := fmt.Sprintf("localhost:%d", port)

	conn, err := grpc.DialContext(context.Background(), addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to javaparser server: %w", err)
	}

	m.conn = conn

	return conn, nil
}

func locateJavaparser() (string, error) {
	runfiles, err := bazel.ListRunfiles()
	if err != nil {
		return "", fmt.Errorf("failed to find javaparser in runfiles: %w", err)
	}

	possiblePaths := make(map[string]bool)
	for _, rf := range runfiles {
		if strings.HasSuffix(rf.ShortPath, javaparser) {
			possiblePaths[rf.Path] = true
		}
	}

	if len(possiblePaths) == 0 {
		return "", fmt.Errorf("failed to find javaparser in runfiles")
	}

	// The set may contain multiple element indirectly pointing to the same thing.
	for path := range possiblePaths {
		return path, nil
	}
	panic("unreachable code")
}

func readPort(path string) (int32, error) {
	startTime := time.Now()
	for {
		if time.Now().Sub(startTime) > 10*time.Second {
			return 0, fmt.Errorf("timed out waiting for port file to be written by javaparser server")
		}

		bs, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return 0, err
		} else {
			portStr := string(bs)
			port, err := strconv.ParseInt(portStr, 10, 32)
			if err != nil {
				return 0, fmt.Errorf("error parsing port (%q) written by javaparser server: %w", portStr, err)
			}
			return int32(port), nil
		}
	}
}

func (m *ServerManager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn == nil {
		return
	}

	cc := pb.NewLifecycleClient(m.conn)

	// If shutdown returns an error, there's really nothing for us to do.
	cc.Shutdown(context.Background(), &pb.ShutdownRequest{})

	m.conn.Close()
	m.conn = nil
}
