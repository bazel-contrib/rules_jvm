package javaparser

import (
	"context"
	"fmt"
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
}

func NewRunner(logger zerolog.Logger, repoRoot string, javaLogLevel string) (*Runner, error) {
	serverManager, err := servermanager.New(repoRoot, javaLogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create javaparser server manager: %v", err)
	}

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
	// Add same-package type references as fully qualified class names.
	// These are simple type names (e.g., "Clock") that Java doesn't require imports for
	// because they're in the same package as the referencing class.
	for _, simpleTypeName := range resp.GetSamePackageTypeReferences() {
		if packageName.Name != "" {
			className := types.NewClassName(packageName, simpleTypeName)
			importedClasses.Add(className)
		}
	}
	exportedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for _, export := range resp.GetExportedClasses() {
		className, err := types.ParseClassName(export)
		if err != nil {
			return nil, fmt.Errorf("failed to parse exports: %w", err)
		}
		exportedClasses.Add(*className)
	}
	importedPackages := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	for _, pkg := range resp.GetImportedPackagesWithoutSpecificClasses() {
		importedPackages.Add(types.NewPackageName(pkg))
	}
	mains := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for _, main := range resp.GetMains() {
		mains.Add(types.NewClassName(packageName, main))
	}
	definedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for _, defined := range resp.GetDefinedClasses() {
		className, err := types.ParseClassName(defined)
		if err != nil {
			return nil, fmt.Errorf("failed to parse defined class %q: %w", defined, err)
		}
		definedClasses.Add(*className)
	}

	return &java.Package{
		Name:                                   packageName,
		ImportedClasses:                        importedClasses,
		ExportedClasses:                        exportedClasses,
		ImportedPackagesWithoutSpecificClasses: importedPackages,
		Mains:                                  mains,
		DefinedClasses:                         definedClasses,
		Files:                                  sorted_set.NewSortedSet(in.Files),
		TestPackage:                            java.IsTestPackage(in.Rel),
		PerClassMetadata:                       perClassMetadata,
	}, nil
}
