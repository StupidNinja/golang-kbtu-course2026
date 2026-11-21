package main

import (
	"errors"
	"fmt"
	"sync"
)

const (
	DefaultWorkerCount   = 2
	DefaultQueueCapacity = 3
)

var (
	ErrQueueFull   = errors.New("task queue is full")
	ErrPoolStopped = errors.New("worker pool is stopped")
)

type WorkerPool struct {
	store   *TaskStore
	queue   chan int
	mu      sync.Mutex
	stopped bool
	workers sync.WaitGroup
}

func NewWorkerPool(store *TaskStore, workerCount, queueCapacity int) (*WorkerPool, error) {
	return newWorkerPool(store, workerCount, queueCapacity, AnalyzeMatches)
}

func newWorkerPool(store *TaskStore, workerCount, queueCapacity int, process func([]Match) (Report, error)) (*WorkerPool, error) {
	if store == nil {
		return nil, errors.New("task store is required")
	}
	if workerCount < 2 {
		return nil, errors.New("worker pool must have at least 2 workers")
	}
	if queueCapacity < 1 {
		return nil, errors.New("queue capacity must be positive")
	}
	if process == nil {
		return nil, errors.New("processor is required")
	}

	pool := &WorkerPool{store: store, queue: make(chan int, queueCapacity)}
	for i := 0; i < workerCount; i++ {
		pool.workers.Add(1)
		go pool.worker(process)
	}
	return pool, nil
}

func (p *WorkerPool) Submit(id int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return ErrPoolStopped
	}
	if err := p.store.markQueued(id); err != nil {
		return err
	}
	select {
	case p.queue <- id:
		return nil
	default:
		p.store.returnToPending(id)
		return ErrQueueFull
	}
}

func (p *WorkerPool) Stop() {
	p.mu.Lock()
	if !p.stopped {
		p.stopped = true
		close(p.queue)
	}
	p.mu.Unlock()
	p.workers.Wait()
}

func (p *WorkerPool) worker(process func([]Match) (Report, error)) {
	defer p.workers.Done()
	for id := range p.queue {
		task, err := p.store.startQueued(id)
		if err != nil {
			continue
		}
		report, processErr := process(task.Input)
		p.store.finishProcessing(id, report, processErr)
	}
}

func (s *TaskStore) startQueued(id int) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task %d not found", id)
	}
	if task.Status != StatusQueued {
		return Task{}, fmt.Errorf("task %d is in state %s", id, task.Status)
	}
	task.Status = StatusProcessing
	s.tasks[id] = task
	return copyTask(task), nil
}

func (s *TaskStore) finishProcessing(id int, report Report, processErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok || task.Status != StatusProcessing {
		return
	}
	if processErr != nil {
		task.Status = StatusFailed
		task.Result = Report{}
	} else {
		task.Status = StatusCompleted
		task.Result = report
	}
	s.tasks[id] = task
}
