package main

import (
	"math"
	"testing"
)

func TestCreateTask(t *testing.T) {
	store := NewTaskStore()
	task, err := store.Create(sampleMatches())
	if err != nil {
		t.Fatal(err)
	}
	if task.ID != 1 {
		t.Errorf("first ID: got %d, want 1", task.ID)
	}
	if task.Status != StatusPending {
		t.Errorf("got status %s, want pending", task.Status)
	}
	if len(task.Input) != 3 || task.Input[0].Kills != 18 {
		t.Errorf("wrong input: %+v", task.Input)
	}
	if task.Result != (Report{}) {
		t.Error("new task should not have a report")
	}
	second, err := store.Create([]Match{{Result: "draw"}})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != 2 {
		t.Errorf("second ID: got %d, want 2", second.ID)
	}
}

func TestCreateInvalid(t *testing.T) {
	tests := []struct {
		name    string
		matches []Match
	}{
		{"nil", nil},
		{"empty", []Match{}},
		{"missing result", []Match{{}}},
		{"unknown result", []Match{{Result: "other"}}},
		{"negative kills", []Match{{Result: "win", Kills: -1}}},
		{"negative deaths", []Match{{Result: "win", Deaths: -1}}},
		{"negative assists", []Match{{Result: "win", Assists: -1}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewTaskStore()
			_, err := store.Create(test.matches)
			if err == nil {
				t.Fatal("expected validation error")
			}
			_, err = store.Get(1)
			if err == nil {
				t.Error("invalid task was saved")
			}
			task, err := store.Create(sampleMatches())
			if err != nil {
				t.Fatal(err)
			}
			if task.ID != 1 {
				t.Error("invalid input consumed an ID")
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	store := NewTaskStore()
	created, err := store.Create(sampleMatches())
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID || got.Status != StatusPending {
		t.Errorf("wrong task: %+v", got)
	}
	if len(got.Input) != 3 || got.Input[0].Kills != 18 {
		t.Errorf("wrong stored input: %+v", got.Input)
	}
}

func TestTaskNotFound(t *testing.T) {
	store := NewTaskStore()
	_, err := store.Get(42)
	if err == nil {
		t.Error("expected Get error for unknown ID")
	}
	_, err = store.Process(42)
	if err == nil {
		t.Error("expected Process error for unknown ID")
	}
}

func TestTaskStart(t *testing.T) {
	task := Task{Status: StatusPending}
	if err := task.start(); err != nil {
		t.Fatal(err)
	}
	if task.Status != StatusProcessing {
		t.Errorf("got status %s, want processing", task.Status)
	}
	if err := task.start(); err == nil {
		t.Error("expected error when starting twice")
	}
	if task.Status != StatusProcessing {
		t.Error("invalid start changed status")
	}
}

func TestProcessTask(t *testing.T) {
	store := NewTaskStore()
	created, err := store.Create(sampleMatches())
	if err != nil {
		t.Fatal(err)
	}
	task, err := store.Process(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != StatusCompleted {
		t.Errorf("got status %s, want completed", task.Status)
	}
	want := Report{MatchCount: 3, Wins: 2, Losses: 1, TotalKills: 48, TotalDeaths: 36,
		TotalAssists: 18, AverageKills: 16, AverageDeaths: 12, AverageAssists: 6}
	if math.Abs(task.Result.WinRate-200.0/3.0) > 1e-9 {
		t.Errorf("wrong win rate: %f", task.Result.WinRate)
	}
	// Процент проверен с допуском; остальные поля здесь сравниваются точно.
	want.WinRate = task.Result.WinRate
	if task.Result != want {
		t.Errorf("got report %+v, want %+v", task.Result, want)
	}
	saved, err := store.Get(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Status != StatusCompleted || saved.Result != task.Result {
		t.Error("completed report was not saved")
	}
}

func TestProcessOverflow(t *testing.T) {
	store := NewTaskStore()
	created, err := store.Create([]Match{{Result: "win", Kills: math.MaxInt}, {Result: "win", Kills: 1}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Process(created.ID)
	if err == nil {
		t.Fatal("expected overflow error")
	}
	saved, err := store.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Status != StatusFailed {
		t.Errorf("got status %s, want failed", saved.Status)
	}
	if saved.Result != (Report{}) {
		t.Error("failed task contains a report")
	}
}

func TestRepeatProcessing(t *testing.T) {
	tests := []struct {
		name    string
		matches []Match
		failed  bool
	}{
		{"completed", sampleMatches(), false},
		{"failed", []Match{{Result: "win", Kills: math.MaxInt}, {Result: "win", Kills: 1}}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewTaskStore()
			created, err := store.Create(test.matches)
			if err != nil {
				t.Fatal(err)
			}
			before, err := store.Process(created.ID)
			if (err != nil) != test.failed {
				t.Fatalf("unexpected first processing error: %v", err)
			}
			_, err = store.Process(created.ID)
			if err == nil {
				t.Fatal("expected error when processing twice")
			}
			after, err := store.Get(created.ID)
			if err != nil {
				t.Fatal(err)
			}
			if after.Status != before.Status || after.Result != before.Result {
				t.Error("repeat processing changed saved task")
			}
		})
	}
}

func TestIndependentTasks(t *testing.T) {
	store := NewTaskStore()
	first, err := store.Create(sampleMatches())
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Create([]Match{{Result: "draw"}})
	if err != nil {
		t.Fatal(err)
	}
	first, err = store.Process(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, err = store.Get(second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if second.Status != StatusPending || second.Result != (Report{}) {
		t.Error("processing first task changed second task")
	}
	second, err = store.Process(second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Error("task IDs must be unique")
	}
	if second.Result != (Report{MatchCount: 1, Draws: 1}) {
		t.Errorf("wrong second report: %+v", second.Result)
	}
	saved, err := store.Get(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Result != first.Result || len(saved.Input) != 3 {
		t.Error("processing second task changed first task")
	}
}

func TestTaskCopies(t *testing.T) {
	store := NewTaskStore()
	input := sampleMatches()
	created, err := store.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0].Kills = 999
	created.Input[1].Kills = 999
	created.Status = StatusCompleted
	saved, err := store.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Input[0].Kills != 18 || saved.Input[1].Kills != 10 || saved.Status != StatusPending {
		t.Error("caller changed task after Create")
	}
	saved.Input[0].Kills = 888
	processed, err := store.Process(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if processed.Result.TotalKills != 48 {
		t.Error("Get returned shared input")
	}
	processed.Input[0].Kills = 777
	processed.Result.TotalKills = 777
	processed.Status = StatusPending
	saved, err = store.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	saved.Result.TotalKills = 666
	again, err := store.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.Input[0].Kills != 18 || again.Result.TotalKills != 48 || again.Status != StatusCompleted {
		t.Error("returned task changed saved input, report or status")
	}
}
