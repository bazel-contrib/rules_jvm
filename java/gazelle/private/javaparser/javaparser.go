package javaparser

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/servermanager"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_multiset"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/status"
)

type Runner struct {
	logger        zerolog.Logger
	rpc           pb.JavaParserClient
	serverManager *servermanager.ServerManager
	prefetch      *prefetcher
}

// NewRunner starts or connects to the javaparser server. At most prefetchConcurrency prefetches
// run at once; 0 turns prefetching off, and a negative value uses min(GOMAXPROCS, 4).
func NewRunner(logger zerolog.Logger, repoRoot string, javaLogLevel string, prefetchConcurrency int) (*Runner, error) {
	serverManager, err := servermanager.New(repoRoot, javaLogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create javaparser server manager: %v", err)
	}

	conn, err := serverManager.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to start / connect to javaparser server: %w", err)
	}

	r := &Runner{
		logger:        logger.With().Str("_c", "javaparser").Logger(),
		rpc:           pb.NewJavaParserClient(conn),
		serverManager: serverManager,
	}
	if prefetchConcurrency < 0 {
		prefetchConcurrency = min(runtime.GOMAXPROCS(0), 4)
	}
	if prefetchConcurrency > 0 {
		r.prefetch = newPrefetcher(r.rpc, prefetchConcurrency)
	}
	return r, nil
}

// Prefetch starts parsing files in rel in the background, for a later ParsePackage call with
// the same Rel and Files. It does nothing if prefetching is off.
func (r *Runner) Prefetch(rel string, files []string) {
	if r.prefetch != nil {
		r.prefetch.schedule(rel, files)
	}
}

// StopPrefetching drops the prefetches that haven't started and waits for the running ones.
func (r *Runner) StopPrefetching() {
	if r.prefetch != nil {
		r.prefetch.stop()
	}
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

	var resp *pb.Package
	var err error
	if r.prefetch != nil {
		resp, err = r.prefetch.parse(ctx, in)
	} else {
		resp, err = r.rpc.ParsePackage(ctx, &pb.ParsePackageRequest{Rel: in.Rel, Files: in.Files})
	}
	if err != nil {
		if grpcErr, ok := status.FromError(err); ok {
			// gRPC is an implementation detail of the javaparser layer, and shouldn't be relied on by higher layers.
			// Reformat gRPC-related details here, for more clear error messages.
			return nil, fmt.Errorf("%s: %s", grpcErr.Code().String(), grpcErr.Message())
		}
		return nil, err
	}

	perClassMetadata := make(map[string]java.PerClassMetadata, len(resp.GetPerClassMetadata()))
	for k, v := range resp.GetPerClassMetadata() {
		annotationClassNames := sorted_set.NewSortedSetFn(nil, types.ClassNameLess)
		for _, annotation := range v.GetAnnotationClassNames() {
			annotationClassName, err := types.ParseClassName(annotation)
			if err != nil {
				return nil, fmt.Errorf("failed to parse annotation name %q as a class name in %s: %w", k, annotation, err)
			}
			annotationClassNames.Add(*annotationClassName)
		}

		methodAnnotationClassNames := sorted_multiset.NewSortedMultiSetFn[string, types.ClassName](types.ClassNameLess)
		for method, perMethod := range v.GetPerMethodMetadata() {
			for _, annotation := range perMethod.AnnotationClassNames {
				annotationClassName, err := types.ParseClassName(annotation)
				if err != nil {
					return nil, fmt.Errorf("failed to parse annotation name %q as a class name in %s: %w", k, annotation, err)
				}
				methodAnnotationClassNames.Add(method, *annotationClassName)
			}
		}
		fieldAnnotationClassNames := sorted_multiset.NewSortedMultiSetFn[string, types.ClassName](types.ClassNameLess)
		for field, perField := range v.GetPerFieldMetadata() {
			for _, annotation := range perField.AnnotationClassNames {
				annotationClassName, err := types.ParseClassName(annotation)
				if err != nil {
					return nil, fmt.Errorf("failed to parse annotation name %q as a class name in %s: %w", k, annotation, err)
				}
				fieldAnnotationClassNames.Add(field, *annotationClassName)
			}
		}
		metadata := java.PerClassMetadata{
			AnnotationClassNames:       annotationClassNames,
			MethodAnnotationClassNames: methodAnnotationClassNames,
			FieldAnnotationClassNames:  fieldAnnotationClassNames,
		}
		perClassMetadata[k] = metadata
	}

	packageName := types.NewPackageName(resp.GetName())
	importedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for _, import_ := range resp.GetImportedClasses() {
		className, err := types.ParseClassName(import_)
		if err != nil {
			return nil, fmt.Errorf("failed to parse imports: %w", err)
		}
		importedClasses.Add(*className)
	}
	exportedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for _, export := range resp.GetExportedClasses() {
		className, err := types.ParseClassName(export)
		if err != nil {
			return nil, fmt.Errorf("failed to parse exports: %w", err)
		}
		exportedClasses.Add(*className)
	}
	internalClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for _, internal := range resp.GetInternalClasses() {
		className, err := types.ParseClassName(internal)
		if err != nil {
			return nil, fmt.Errorf("failed to parse internal classes: %w", err)
		}
		internalClasses.Add(*className)
	}
	declaredClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for _, declared := range resp.GetDeclaredClasses() {
		className, err := types.ParseClassName(declared)
		if err != nil {
			return nil, fmt.Errorf("failed to parse declared classes: %w", err)
		}
		declaredClasses.Add(*className)
	}
	importedPackages := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	for _, pkg := range resp.GetImportedPackagesWithoutSpecificClasses() {
		importedPackages.Add(types.NewPackageName(pkg))
	}
	mains := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for _, main := range resp.GetMains() {
		mains.Add(types.NewClassName(packageName, main))
	}

	return &java.Package{
		Name:                                   packageName,
		ImportedClasses:                        importedClasses,
		ExportedClasses:                        exportedClasses,
		InternalClasses:                        internalClasses,
		DeclaredClasses:                        declaredClasses,
		ImportedPackagesWithoutSpecificClasses: importedPackages,
		Mains:                                  mains,
		Files:                                  sorted_set.NewSortedSet(in.Files),
		TestPackage:                            java.IsTestPackage(in.Rel),
		PerClassMetadata:                       perClassMetadata,
	}, nil
}
