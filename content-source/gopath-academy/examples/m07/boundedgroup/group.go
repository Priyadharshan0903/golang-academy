package boundedgroup

import (
	"context"
	"sync"
)

// [start group]

// Group runs functions in goroutines, cancels the shared context on
// the first error, and returns that error from Wait. It is a small
// version of golang.org/x/sync/errgroup, built only on the standard
// library, to show how that package works.
type Group struct {
	cancel  context.CancelCauseFunc
	wg      sync.WaitGroup
	sem     chan struct{} // nil means "no limit"
	errOnce sync.Once
	err     error
}

// WithContext returns a Group and a context that is cancelled when
// any function returns an error, or when Wait returns.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}

// SetLimit allows at most n functions to run at once.
func (g *Group) SetLimit(n int) { g.sem = make(chan struct{}, n) }

// Go runs f in a new goroutine, blocking first if the limit is reached.
func (g *Group) Go(f func() error) {
	if g.sem != nil {
		g.sem <- struct{}{} // acquire a slot (blocks when full)
	}
	g.wg.Add(1)
	go func() {
		defer func() {
			if g.sem != nil {
				<-g.sem // release the slot
			}
			g.wg.Done()
		}()
		if err := f(); err != nil {
			g.errOnce.Do(func() { // keep only the FIRST error
				g.err = err
				if g.cancel != nil {
					g.cancel(err)
				}
			})
		}
	}()
}

// Wait blocks until all functions return, then returns the first error.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}

// [end group]
