package servermanager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"google.golang.org/grpc"
)

type ServerManager struct {
	workspace    string
	javaLogLevel string
	tmpdir       string

	mu   sync.Mutex
	conn *grpc.ClientConn
}

type JavaparserLocator interface {
	locateJavaparser(jvmFlags []string) (string, error)
	startupFlags(jvmFlags []string) []string
}

func New(workspace, javaLogLevel string) (*ServerManager, error) {
	dir, err := os.MkdirTemp(os.Getenv("TMPDIR"), "gazelle-javaparser")
	if err != nil {
		return nil, fmt.Errorf("failed to create tmpdir to start javaparser server: %w", err)
	}
	return &ServerManager{
		workspace:    workspace,
		javaLogLevel: javaLogLevel,
		tmpdir:       dir,
	}, nil
}

func (m *ServerManager) Connect() (*grpc.ClientConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil {
		return m.conn, nil
	}
	logLevelFlag := fmt.Sprintf("-Dorg.slf4j.simpleLogger.defaultLogLevel=%s", m.javaLogLevel)

	jvmFlags := []string{
		logLevelFlag,
	}

	portFilePath := filepath.Join(m.tmpdir, "port")

	javaParserPath, err := m.locateJavaparser(jvmFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to find javaparser: %w", err)
	}

	commandArgs := []string{
		javaParserPath,
	}
	for _, arg := range m.startupFlags(jvmFlags) {
		commandArgs = append(commandArgs, arg)
	}
	commandArgs = append(commandArgs,
		"--server-port-file-path", portFilePath,
		"--workspace", m.workspace,
		"--idle-timeout", "30",
	)

	cmd := exec.Command(commandArgs[0], commandArgs[1:]...)

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

	// Ask the server to shut down, but don't block indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _ = cc.Shutdown(ctx, &pb.ShutdownRequest{})

	// Give the server a brief moment to send GOAWAY and close cleanly to avoid DEBUG stacktraces.
	time.Sleep(100 * time.Millisecond)

	m.conn.Close()

	m.conn = nil
}
