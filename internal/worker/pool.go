package worker

import (
	"context"
	"log"
	"sync"

	"github.com/danlju/tulip-worker/internal/queue"
)

type Job struct {
	Msg queue.Message
}

type Handler interface {
	Handle(ctx context.Context, msg queue.Message) error
}

type Pool struct {
	jobs   chan Job
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewPool(size int, handler Handler) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Pool{
		jobs:   make(chan Job, size*2),
		ctx:    ctx,
		cancel: cancel,
	}

	for i := 0; i < size; i++ {
		p.wg.Add(1)
		go p.worker(i, handler)
	}

	return p
}

func (p *Pool) worker(id int, handler Handler) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("worker %d shutting down", id)
			return

		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			if err := handler.Handle(p.ctx, job.Msg); err != nil {
				log.Printf("worker %d error: %v", id, err)
			}
		}
	}
}

func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

func (p *Pool) Available() int {
	return cap(p.jobs) - len(p.jobs)
}

func (p *Pool) Shutdown() {
	p.cancel()
	close(p.jobs)
	p.wg.Wait()
}
