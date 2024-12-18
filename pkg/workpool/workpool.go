package workpool

import (
	"context"
	"sync"
)

type Result struct {
	Msg string
	Err error
}

type Pool struct {
	wg      sync.WaitGroup
	jobs    chan func(ctx context.Context) Result
	Results chan Result
}

func NewPool(ctx context.Context, nWorkers int) *Pool {
	p := &Pool{
		wg:      sync.WaitGroup{},
		jobs:    make(chan func(ctx context.Context) Result, nWorkers),
		Results: make(chan Result, nWorkers),
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

	return p
}

func (p *Pool) Schedule(ctx context.Context, job func(ctx context.Context) Result) {
	select {
	case p.jobs <- job:
		break
	case <-ctx.Done():
		break
	}
}

func (p *Pool) Wait() {
	p.wg.Wait()
	close(p.Results)
}
