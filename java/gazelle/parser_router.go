package gazelle

import (
	"context"
	"strings"
	"sync"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/parser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/tsparser"
	"github.com/rs/zerolog"
)

type parserRouter struct {
	javaParser *tsparser.Runner

	logger       zerolog.Logger
	repoRoot     string
	javaLogLevel string

	mu             sync.Mutex
	javacParser    parser.Parser
	javacParserErr error
}

func newParserRouter(logger zerolog.Logger, repoRoot, javaLogLevel string) parser.Parser {
	return &parserRouter{
		javaParser:   tsparser.NewRunner(logger, repoRoot),
		logger:       logger,
		repoRoot:     repoRoot,
		javaLogLevel: javaLogLevel,
	}
}

func (p *parserRouter) ParsePackage(ctx context.Context, in *parser.ParsePackageRequest) (*java.Package, error) {
	for _, filename := range in.Files {
		if strings.HasSuffix(filename, ".kt") {
			javacParser, err := p.kotlinParser()
			if err != nil {
				return nil, err
			}
			return javacParser.ParsePackage(ctx, in)
		}
	}
	return p.javaParser.ParsePackage(ctx, in)
}

func (p *parserRouter) Shutdown() {
	p.javaParser.Shutdown()
	p.mu.Lock()
	javacParser := p.javacParser
	p.mu.Unlock()
	shutdownParser(javacParser)
}

func (p *parserRouter) kotlinParser() (parser.Parser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.javacParser != nil || p.javacParserErr != nil {
		return p.javacParser, p.javacParserErr
	}
	p.javacParser, p.javacParserErr = javaparser.NewRunner(p.logger, p.repoRoot, p.javaLogLevel)
	return p.javacParser, p.javacParserErr
}
