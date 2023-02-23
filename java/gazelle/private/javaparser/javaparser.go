package javaparser

import (
	"context"
	"fmt"
	"time"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/servermanager"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/rs/zerolog"
)

type Runner struct {
	logger        zerolog.Logger
	rpc           pb.JavaParserClient
	serverManager *servermanager.ServerManager
}

func NewRunner(logger zerolog.Logger, repoRoot string, javaLogLevel string) (*Runner, error) {
	serverManager := servermanager.New(repoRoot, javaLogLevel)

	conn, err := serverManager.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to start / connect to javaparser server: %w", err)
	}

	return &Runner{
		logger:        logger.With().Str("_c", "javaparser").Logger(),
		rpc:           pb.NewJavaParserClient(conn),
		serverManager: serverManager,
	}, nil
}

func (r *Runner) ServerManager() *servermanager.ServerManager {
	return r.serverManager
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

	perClassMetadata := make(map[string]java.PerClassMetadata, len(resp.GetPerClassMetadata()))
	for k, v := range resp.GetPerClassMetadata() {
		metadata := java.PerClassMetadata{
			AnnotationClassNames: sorted_set.NewSortedSet(v.GetAnnotationClassNames()),
		}
		perClassMetadata[k] = metadata
	}

	return &java.Package{
		Name:                                   resp.GetName(),
		ImportedClasses:                        sorted_set.NewSortedSet(resp.GetImportedClasses()),
		ImportedPackagesWithoutSpecificClasses: sorted_set.NewSortedSet(resp.GetImportedPackagesWithoutSpecificClasses()),
		Mains:                                  sorted_set.NewSortedSet(resp.GetMains()),
		Files:                                  sorted_set.NewSortedSet(in.Files),
		TestPackage:                            java.IsTestPath(in.Rel),
		PerClassMetadata:                       perClassMetadata,
	}, nil
}
