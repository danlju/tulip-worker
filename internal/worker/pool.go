package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/danlju/tulip-worker/internal/queue"
)

type Job struct {
	Msg       queue.Message
	Ack       func(context.Context) error
	Heartbeat func(context.Context) error
}

type Handler interface {
	Handle(ctx context.Context, msg queue.Message) error
}

type Pool struct {
	jobs      chan Job
	available chan struct{}
	wg        sync.WaitGroup
}

func NewPool(size int, handler Handler) *Pool {
	p := &Pool{
		jobs:      make(chan Job, size),
		available: make(chan struct{}, size),
	}
	for i := 0; i < size; i++ {
		p.available <- struct{}{}
	}

	for i := 0; i < size; i++ {
		p.wg.Add(1)
		go p.worker(i, handler)
	}

	return p
}

func (p *Pool) worker(id int, handler Handler) {
	defer p.wg.Done()

	for job := range p.jobs {
		ctx := context.Background()
		heartbeatDone := make(chan struct{})
		if job.Heartbeat != nil {
			go func() {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-heartbeatDone:
						return
					case <-ticker.C:
						if err := job.Heartbeat(ctx); err != nil {
							log.Printf("worker %d visibility extension failed: %v", id, err)
						}
					}
				}
			}()
		}
		err := handler.Handle(ctx, job.Msg)
		close(heartbeatDone)
		if err != nil {
			log.Printf("worker %d error; leaving message for retry: %v", id, err)
		} else if job.Ack != nil {
			if err := job.Ack(ctx); err != nil {
				log.Printf("worker %d failed to acknowledge message: %v", id, err)
			}
		}
		p.available <- struct{}{}
	}
}

func (p *Pool) Submit(job Job) {
	<-p.available
	p.jobs <- job
}

func (p *Pool) Available() int {
	return len(p.available)
}

func (p *Pool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
}
