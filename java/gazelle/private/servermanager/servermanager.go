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
	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

const (
	javaparserPath = "contrib_rules_jvm/java/src/com/github/bazel_contrib/contrib_rules_jvm/javaparser/generators/Main"

	runfilesManifestFileKey = "RUNFILES_MANIFEST_FILE"
	runfilesDirKey          = "RUNFILES_DIR"
	javaRunfilesKey         = "JAVA_RUNFILES"
)

type ServerManager struct {
	logger       zerolog.Logger
	workspace    string
	javaLogLevel string

	mu   sync.Mutex
	conn *grpc.ClientConn
}

func New(workspace, javaLogLevel string, logger zerolog.Logger) *ServerManager {
	return &ServerManager{
		workspace:    workspace,
		javaLogLevel: javaLogLevel,
		logger:       logger,
	}
}

func (m *ServerManager) Connect() (*grpc.ClientConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil {
		return m.conn, nil
	}

	logLevelFlag := fmt.Sprintf("-Dorg.slf4j.simpleLogger.defaultLogLevel=%s", m.javaLogLevel)

	javaParserPath, err := m.locateJavaparser()
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
		"--idle-timeout", "30",
	)

	runfilesEnv, err := m.runfilesEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get runfiles env vars for javaparser: %w", err)
	}

	cmd.Env = append(cmd.Env, runfilesEnv...)

	m.logger.Debug().
		Str("cmd", cmd.String()).
		Strs("env", cmd.Env).
		Msg("Starting javaparser with command")

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

func (m *ServerManager) locateJavaparser() (string, error) {
	path, err := runfiles.Rlocation(javaparserPath)
	if err != nil {
		return "", fmt.Errorf("failed to find javaparser in runfiles: %w", err)
	}

	return path, nil
}

func (m *ServerManager) runfilesEnv() ([]string, error) {
	envVars, err := runfiles.Env()
	if err != nil {
		return nil, fmt.Errorf("failed to get runfiles env: %w", err)
	}

	m.logger.Debug().
		Strs("res", envVars).
		Msg("resolved runfiles env")

	var (
		hasJavaRunfiles  bool
		runfilesManifest string
	)

	res := make([]string, 0, len(envVars)+2) // NOTE: potentially adding two more env vars.

	for _, envVar := range envVars {
		if runfilesManifest == "" && strings.HasPrefix(envVar, runfilesManifestFileKey) {
			runfilesManifest = strings.ReplaceAll(envVar, runfilesManifestFileKey+"=", "")
		}

		if !hasJavaRunfiles && strings.HasPrefix(envVar, runfilesDirKey) {
			hasJavaRunfiles = true
			path := strings.ReplaceAll(envVar, javaRunfilesKey+"=", "")
			res = append(res, javaRunfilesKey+"="+path)
		}

		if !hasJavaRunfiles && strings.HasPrefix(envVar, javaRunfilesKey) {
			hasJavaRunfiles = true
		}

		res = append(res, envVar)
	}

	if !hasJavaRunfiles && runfilesManifest != "" {
		manifestDirPath := strings.ReplaceAll(runfilesManifest, runfilesManifestFileKey+"=", "")
		runfilesDirPath := strings.ReplaceAll(manifestDirPath, "_manifest", "")

		res = append(res, javaRunfilesKey+"="+runfilesDirPath)
	}

	return res, nil
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
