package workpool

import (
	"context"
	"sync"
)

type Pool struct {
	wg      sync.WaitGroup
	jobs    chan func(ctx context.Context) error
	Results chan error
}

func NewPool(ctx context.Context, nWorkers int) *Pool {
	p := &Pool{
		wg:      sync.WaitGroup{},
		jobs:    make(chan func(ctx context.Context) error, nWorkers),
		Results: make(chan error, nWorkers),
	}

	p.wg.Add(nWorkers)

	for i := 0; i < nWorkers; i++ {
		go func() {
			defer p.wg.Done()

			for {
				select {
				case job := <-p.jobs:
					p.Results <- job(ctx)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		p.wg.Wait()
		close(p.Results)
	}()

	return p
}

func (p *Pool) Schedule(job func(ctx context.Context) error) {
	p.jobs <- job
}
