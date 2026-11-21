package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	err := run(os.Args[1:], os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(paths []string, out io.Writer) error {
	if len(paths) == 0 {
		return errors.New("usage: go run . session.json [another-session.json ...]")
	}
	store := NewTaskStore()
	var tasks []Task
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("%s: %v", path, err)
		}
		matches, err := DecodeSession(file)
		closeErr := file.Close()
		if err != nil {
			return fmt.Errorf("%s: %v", path, err)
		}
		if closeErr != nil {
			return fmt.Errorf("%s: %v", path, closeErr)
		}

		task, err := store.Create(matches)
		if err != nil {
			return fmt.Errorf("%s: %v", path, err)
		}
		_, err = store.Process(task.ID)
		if err != nil {
			return fmt.Errorf("%s: %v", path, err)
		}
		task, err = store.Get(task.ID)
		if err != nil {
			return fmt.Errorf("%s: %v", path, err)
		}
		tasks = append(tasks, task)
	}

	// Печатаем после проверки всех файлов, чтобы не выводить неполный результат.
	for _, task := range tasks {
		err := writeReport(out, task)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeReport(w io.Writer, task Task) error {
	if task.Status != StatusCompleted {
		return fmt.Errorf("task %d has no completed report", task.ID)
	}
	report := task.Result
	_, err := fmt.Fprintf(w, `Task: %d
Status: %s
Matches: %d
Wins: %d
Losses: %d
Draws: %d
Total kills: %d
Total deaths: %d
Total assists: %d
Win rate: %.2f%%
Average kills: %.2f
Average deaths: %.2f
Average assists: %.2f

`, task.ID, task.Status, report.MatchCount, report.Wins, report.Losses, report.Draws,
		report.TotalKills, report.TotalDeaths, report.TotalAssists, report.WinRate,
		report.AverageKills, report.AverageDeaths, report.AverageAssists)
	return err
}
