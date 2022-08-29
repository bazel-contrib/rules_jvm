package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/bazel"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/cmd/javaparser-wrapper/internal/activitytracker"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/netutil"
	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// runnerTimout is the time after which the runner will die.
const runnerTimout = 30 * time.Second

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var addr, workspacePath string
	var jvmArgs arrayFlags

	flag.StringVar(&addr, "addr", "", "Address to listen")
	flag.StringVar(&workspacePath, "workspace", "", "Where is the workspace")
	flag.Var(&jvmArgs, "jvm_arg", "Pass <flag> to the java command itself. <flag> may contain spaces. Can be used multiple times.")
	flag.Parse()

	javaLevel := "info"
	for _, arg := range jvmArgs {
		if strings.HasPrefix(arg, "-Dorg.slf4j.simpleLogger.defaultLogLevel=") {
			javaLevel = strings.TrimPrefix(arg, "-Dorg.slf4j.simpleLogger.defaultLogLevel=")
		}
	}

	goLevel, err := zerolog.ParseLevel(javaLevel)
	if err != nil {
		panic(err.Error())
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(goLevel)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if workspacePath == "" {
		logger.Error().Msg("missing workspace parameter")
		flag.Usage()
		os.Exit(1)
	}

	binPath, err := findBinary(logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not find BFG")
	}

	realBFGAddr, err := netutil.RandomAddr()
	if err != nil {
		logger.Fatal().Msg("could not get an adresse for BFG")
	}

	_, port, err := net.SplitHostPort(realBFGAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not get port from address")
	}

	var args []string
	for _, arg := range jvmArgs {
		args = append(args, "--jvm_flag="+arg)
	}
	args = append(
		args,
		"--server-port", port,
		"--workspace", workspacePath,
	)

	cmd := exec.Command(binPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logger.Debug().Strs("args", cmd.Args).Msg("bfg-server args")
	if err := cmd.Start(); err != nil {
		logger.Fatal().Err(err).Msg("could not start command")
	}
	defer func() {
		logger.Debug().Msg("stopping real BFG")
		if err := cmd.Process.Kill(); err != nil {
			logger.Error().Err(err).Msg("failed to stop real BFG")
		}
	}()

	conn, err := grpc.Dial(realBFGAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Fatal().Err(err).Msg("did not connect")
	}
	defer conn.Close()

	srv := newServer(logger, workspacePath, pb.NewJavaParserClient(conn))
	at := activitytracker.NewTracker()
	s := grpc.NewServer(grpc.UnaryInterceptor(activityTrackerUSI(at)))
	pb.RegisterJavaParserServer(s, srv)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to listen")
	}
	defer ln.Close()

	logger.Debug().Stringer("addr", ln.Addr()).Msg("javaparser-wrapper server listening")

	go func() {
		if err := s.Serve(ln); err != nil {
			logger.Fatal().Err(err).Msg("failed to serve")
		}
	}()

	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case ts := <-t.C:
				logger.Debug().Str("ts", ts.String()).Msg("reaper ticked")
				past := ts.Add(-runnerTimout)
				if at.IdleSince(past) {
					logger.Debug().Str("ts", ts.String()).Str("past", past.String()).Msg("reaper stopping process")
					cancel()
				}
			}
		}
	}()

	go func() {
		oscall := <-c
		logger.Debug().Str("syscall", fmt.Sprintf("%+v", oscall)).Msg("syscall")
		cancel()
	}()

	<-ctx.Done()
	logger.Debug().Msg("graceful stop started")
	s.GracefulStop()
	logger.Debug().Msg("graceful stop done")
}

func activityTrackerUSI(at *activitytracker.Tracker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		at.Ping()
		return handler(ctx, req)
	}
}

type server struct {
	logger    zerolog.Logger
	real      pb.JavaParserClient
	workspace string
}

func newServer(logger zerolog.Logger, workspace string, real pb.JavaParserClient) *server {
	return &server{
		logger:    logger.With().Str("_c", "server").Logger(),
		real:      real,
		workspace: workspace,
	}
}

func (s *server) ParsePackage(ctx context.Context, in *pb.ParsePackageRequest) (*pb.Package, error) {
	s.logger.Debug().
		Str("name", "ParsePackage").
		Msg("RPC")

	pkg, err := s.real.ParsePackage(ctx, in)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

// findBinary finds the build_file_generator.
//
// forked and simplified from bazel.FindBinary.
func findBinary(logger zerolog.Logger) (string, error) {
	entries, err := bazel.ListRunfiles()
	if err != nil {
		return "", err
	}

	wValues := make(map[string]bool)

	for _, entry := range entries {
		if _, ok := wValues[entry.Workspace]; !ok {
			logger.Debug().Msgf("new workspace: %#v", entry.Workspace)
			wValues[entry.Workspace] = true
		}
		if entry.Workspace != "contrib_rules_jvm" {
			continue
		}
		if entry.ShortPath != "java/src/com/github/bazel_contrib/contrib_rules_jvm/javaparser/generators/Main" {
			continue
		}
		logger.Debug().Msgf("entry: %#v", entry)
		return entry.Path, nil
	}

	return "", fmt.Errorf("could not find javaparser")
}
