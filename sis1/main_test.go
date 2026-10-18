package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleSessionJSON = `[
  {"result":"win","kills":18,"deaths":12,"assists":6},
  {"result":"loss","kills":10,"deaths":16,"assists":4},
  {"result":"win","kills":20,"deaths":8,"assists":8}
]`

func TestWriteReport(t *testing.T) {
	task := Task{
		ID:     7,
		Status: StatusCompleted,
		Result: Report{
			MatchCount: 3, Wins: 2, Losses: 1,
			TotalKills: 48, TotalDeaths: 36, TotalAssists: 18,
			WinRate: 200.0 / 3.0, AverageKills: 16, AverageDeaths: 12, AverageAssists: 6,
		},
	}
	var output strings.Builder
	err := writeReport(&output, task)
	if err != nil {
		t.Fatal(err)
	}
	want := `Task: 7
Status: completed
Matches: 3
Wins: 2
Losses: 1
Draws: 0
Total kills: 48
Total deaths: 36
Total assists: 18
Win rate: 66.67%
Average kills: 16.00
Average deaths: 12.00
Average assists: 6.00

`
	if output.String() != want {
		t.Errorf("got report:\n%s\nwant:\n%s", output.String(), want)
	}
}

func TestWriteReportInvalidState(t *testing.T) {
	for _, status := range []string{StatusPending, StatusProcessing, StatusFailed} {
		t.Run(status, func(t *testing.T) {
			var output strings.Builder
			err := writeReport(&output, Task{ID: 1, Status: status})
			if err == nil {
				t.Error("expected error for unfinished task")
			}
			if output.Len() != 0 {
				t.Error("unfinished report was printed")
			}
		})
	}
}

func TestWriteReportError(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	err = writeReport(file, Task{ID: 1, Status: StatusCompleted})
	if err == nil {
		t.Error("expected error from closed output file")
	}
}

func TestRunMultipleSessions(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "session.json")
	second := filepath.Join(dir, "draw.json")
	if err := os.WriteFile(first, []byte(sampleSessionJSON), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte(`[{"result":"draw","kills":0,"deaths":0,"assists":0}]`), 0600); err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	err := run([]string{first, second}, &output)
	if err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if strings.Count(text, "Task:") != 2 {
		t.Errorf("expected two reports, got:\n%s", text)
	}
	if !strings.Contains(text, "Task: 1\nStatus: completed\nMatches: 3\nWins: 2\nLosses: 1\nDraws: 0") {
		t.Error("first report is missing or incorrect")
	}
	if !strings.Contains(text, "Total kills: 48\nTotal deaths: 36\nTotal assists: 18\nWin rate: 66.67%") {
		t.Error("first report has wrong statistics")
	}
	if !strings.Contains(text, "Task: 2\nStatus: completed\nMatches: 1\nWins: 0\nLosses: 0\nDraws: 1") {
		t.Error("second report is missing or incorrect")
	}
	if !strings.Contains(text, "Win rate: 0.00%\nAverage kills: 0.00\nAverage deaths: 0.00\nAverage assists: 0.00") {
		t.Error("draw report has wrong statistics")
	}
}

func TestRunInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		noArgs  bool
		missing bool
	}{
		{name: "no arguments", noArgs: true},
		{name: "missing file", missing: true},
		{name: "broken JSON", input: `[`},
		{name: "invalid stats", input: `[{"result":"win","kills":-1,"deaths":0,"assists":0}]`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.json")
			args := []string{path}
			if test.noArgs {
				args = nil
			}
			if !test.noArgs && !test.missing {
				if err := os.WriteFile(path, []byte(test.input), 0600); err != nil {
					t.Fatal(err)
				}
			}
			var output strings.Builder
			err := run(args, &output)
			if err == nil {
				t.Fatal("expected run error")
			}
			if !test.noArgs && !strings.Contains(err.Error(), path) {
				t.Errorf("file path missing in error: %v", err)
			}
			if output.Len() != 0 {
				t.Error("invalid run printed a report")
			}
		})
	}
}

func TestRunStopsBeforeOutput(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "valid.json")
	second := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(first, []byte(sampleSessionJSON), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte(`[]`), 0600); err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	if err := run([]string{first, second}, &output); err == nil {
		t.Fatal("expected invalid second session error")
	}
	if output.Len() != 0 {
		t.Error("partial results printed before all inputs succeeded")
	}
}
