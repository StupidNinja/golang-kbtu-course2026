package main

import "fmt"

const (
	StatusPending    = "pending"
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
	task, ok := s.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task %d not found", id)
	}
	err := task.start()
	if err != nil {
		return copyTask(task), err
	}
	s.tasks[id] = task

	report, err := AnalyzeMatches(task.Input)
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

func copyTask(task Task) Task {
	// Копия структуры всё ещё делит массив Input, поэтому копируем срез отдельно.
	matches := make([]Match, len(task.Input))
	copy(matches, task.Input)
	task.Input = matches
	return task
}
