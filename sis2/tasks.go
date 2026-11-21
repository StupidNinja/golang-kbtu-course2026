package main

import (
	"fmt"
	"sync"
)

const (
	StatusPending    = "pending"
	StatusQueued     = "queued"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

type Task struct {
	ID     int
	Input  []Match
	Status string
	Result Report
}

type TaskStore struct {
	mu     sync.RWMutex
	tasks  map[int]Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

func (s *TaskStore) Create(matches []Match) (Task, error) {
	err := validateMatches(matches)
	if err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task := Task{
		ID:     s.nextID,
		Input:  matches,
		Status: StatusPending,
	}
	s.tasks[task.ID] = copyTask(task)
	s.nextID++
	return copyTask(task), nil
}

func (s *TaskStore) Get(id int) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task %d not found", id)
	}
	return copyTask(task), nil
}

func (t *Task) start() error {
	if t.Status != StatusPending {
		return fmt.Errorf("task %d cannot be processed in state %s", t.ID, t.Status)
	}
	t.Status = StatusProcessing
	return nil
}

func (s *TaskStore) Process(id int) (Task, error) {
	return s.process(id, StatusPending)
}

func (s *TaskStore) process(id int, expectedStatus string) (Task, error) {
	s.mu.Lock()
	task, ok := s.tasks[id]
	if !ok {
		s.mu.Unlock()
		return Task{}, fmt.Errorf("task %d not found", id)
	}
	if task.Status != expectedStatus {
		status := task.Status
		s.mu.Unlock()
		return copyTask(task), fmt.Errorf("task %d cannot be processed in state %s", id, status)
	}
	task.Status = StatusProcessing
	s.tasks[id] = task
	s.mu.Unlock()

	report, err := AnalyzeMatches(task.Input)
	s.mu.Lock()
	defer s.mu.Unlock()
	task = s.tasks[id]
	if err != nil {
		task.Status = StatusFailed
		task.Result = Report{}
		s.tasks[id] = task
		return copyTask(task), err
	}
	task.Result = report
	task.Status = StatusCompleted
	s.tasks[id] = task
	return copyTask(task), nil
}

func (s *TaskStore) markQueued(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("task %d not found", id)
	}
	if task.Status != StatusPending {
		return fmt.Errorf("task %d cannot be submitted in state %s", id, task.Status)
	}
	task.Status = StatusQueued
	s.tasks[id] = task
	return nil
}

func (s *TaskStore) returnToPending(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task := s.tasks[id]
	if task.Status == StatusQueued {
		task.Status = StatusPending
		s.tasks[id] = task
	}
}

func copyTask(task Task) Task {
	// Копия структуры всё ещё делит массив Input, поэтому копируем срез отдельно.
	matches := make([]Match, len(task.Input))
	copy(matches, task.Input)
	task.Input = matches
	return task
}
