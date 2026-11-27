package javaparser

import (
	"context"
	"slices"
	"sync"

	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
)

// prefetcher sends ParsePackage RPCs ahead of the GenerateRules calls that need them, so the
// javaparser server parses several packages at once instead of one at a time.
//
// A prefetched response is only used for a request with the same Rel and Files. A prefetch
// that hasn't started when it is needed is dropped and the caller sends the request itself,
// outside of the concurrency limit.
type prefetcher struct {
	rpc    pb.JavaParserClient
	ctx    context.Context
	cancel context.CancelFunc
	sem    chan struct{}
	wg     sync.WaitGroup

	mu      sync.Mutex
	stopped bool
	pending map[string]*prefetchedParse
}

type prefetchedParse struct {
	files []string
	// started is set, under prefetcher.mu, when the RPC is sent; done is closed when resp
	// and err are set. A prefetch only starts while it is still in prefetcher.pending.
	started bool
	done    chan struct{}
	resp    *pb.Package
	err     error
}

// newPrefetcher returns a prefetcher that runs at most concurrency prefetches at a time.
// concurrency must be positive.
func newPrefetcher(rpc pb.JavaParserClient, concurrency int) *prefetcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &prefetcher{
		rpc:     rpc,
		ctx:     ctx,
		cancel:  cancel,
		sem:     make(chan struct{}, concurrency),
		pending: make(map[string]*prefetchedParse),
	}
}

// schedule starts parsing files in rel in the background.
func (p *prefetcher) schedule(rel string, files []string) {
	if len(files) == 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	e := &prefetchedParse{files: files, done: make(chan struct{})}
	p.pending[rel] = e
	p.wg.Add(1)
	go p.run(rel, e)
}

func (p *prefetcher) run(rel string, e *prefetchedParse) {
	defer p.wg.Done()
	select {
	case p.sem <- struct{}{}:
	case <-p.ctx.Done():
		return
	}
	defer func() { <-p.sem }()

	p.mu.Lock()
	if p.pending[rel] != e {
		p.mu.Unlock()
		return
	}
	e.started = true
	p.mu.Unlock()

	// Not cancelled by stop: the javaparser server logs a stack trace for each cancelled call.
	e.resp, e.err = p.rpc.ParsePackage(context.Background(), &pb.ParsePackageRequest{Rel: rel, Files: e.files})
	close(e.done)
}

// parse returns the response to in, from a prefetch if one with the same request is running
// or finished, or else by sending the request.
func (p *prefetcher) parse(ctx context.Context, in *ParsePackageRequest) (*pb.Package, error) {
	p.mu.Lock()
	e := p.pending[in.Rel]
	delete(p.pending, in.Rel)
	if e != nil && (!e.started || !slices.Equal(e.files, in.Files)) {
		// Not started, so it never will now that it isn't pending, or a different request.
		e = nil
	}
	p.mu.Unlock()

	if e == nil {
		return p.rpc.ParsePackage(ctx, &pb.ParsePackageRequest{Rel: in.Rel, Files: in.Files})
	}
	select {
	case <-e.done:
		return e.resp, e.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// stop cancels the prefetches that haven't started, and waits for the running ones to finish.
// Prefetch requests made after stop are ignored.
func (p *prefetcher) stop() {
	p.mu.Lock()
	p.stopped = true
	p.pending = make(map[string]*prefetchedParse)
	p.mu.Unlock()
	p.cancel()
	p.wg.Wait()
}
