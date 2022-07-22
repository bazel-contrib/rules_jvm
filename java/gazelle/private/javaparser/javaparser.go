package javaparser

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/bazel"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/netutil"
	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Runner struct {
	logger   zerolog.Logger
	repoRoot string

	rpc        pb.JavaParserClient
	rpcProcess *os.Process
	conn       *grpc.ClientConn
}

func NewRunner(logger zerolog.Logger, repoRoot string, javaLogLevel string) *Runner {
	wrapperPath, found := bazel.FindBinary("java/gazelle/private/javaparser/cmd/javaparser-wrapper", "javaparser-wrapper")
	if !found {
		logger.Fatal().Msg("could not find javaparser-wrapper")
	}

	r := Runner{
		logger:   logger.With().Str("_c", "javaparser-runnerwrapper").Logger(),
		repoRoot: repoRoot,
	}

	addr, err := netutil.RandomAddr()
	if err != nil {
		logger.Fatal().Msg("could not get an address for javaparser-wrapper")
	}

	cmd := exec.Command(
		wrapperPath,
		"--jvm_arg=-Dorg.slf4j.simpleLogger.defaultLogLevel="+javaLogLevel,
		"--server",
		"--addr="+addr,
		"--workspace="+repoRoot,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	r.logger.Debug().Strs("args", cmd.Args).Msg("javaparser args")
	if err := cmd.Start(); err != nil {
		r.logger.Fatal().Err(err).Msg("could not start command")
	}

	// FIXME we do not have a destructor of the language to stop this
	r.logger.Debug().
		Int("pid", cmd.Process.Pid).
		Msg("gazelle does not know how to kill javaparser, do it yourself")
	r.rpcProcess = cmd.Process

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(30*time.Second))
	if err != nil {
		r.logger.Fatal().Err(err).Msg("error connecting to javaparser-wrapper")
	}
	r.conn = conn
	r.rpc = pb.NewJavaParserClient(r.conn)

	return &r
}

type ParsePackageRequest struct {
	Rel   string
	Files []string
}

func (r Runner) ParsePackage(ctx context.Context, in *ParsePackageRequest) (*java.Package, error) {
	defer func(t time.Time) {
		r.logger.Debug().
			Str("duration", time.Since(t).String()).
			Str("rel", in.Rel).
			Msg("parse package done")
	}(time.Now())

	resp, err := r.rpc.ParsePackage(ctx, &pb.ParsePackageRequest{Rel: in.Rel, Files: in.Files})
	if err != nil {
		return nil, err
	}

	return &java.Package{
		Name:        resp.GetName(),
		Imports:     sorted_set.NewSortedSet(resp.GetImports()),
		Mains:       sorted_set.NewSortedSet(resp.GetMains()),
		Files:       sorted_set.NewSortedSet(in.Files),
		TestPackage: java.IsTestPath(in.Rel),
	}, nil
}
