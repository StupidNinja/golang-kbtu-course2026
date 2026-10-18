package main

import (
	"math"
	"testing"
)

func sampleMatches() []Match {
	return []Match{
		{Result: "win", Kills: 18, Deaths: 12, Assists: 6},
		{Result: "loss", Kills: 10, Deaths: 16, Assists: 4},
		{Result: "win", Kills: 20, Deaths: 8, Assists: 8},
	}
}

func TestAnalyzeMatches(t *testing.T) {
	report, err := AnalyzeMatches(sampleMatches())
	if err != nil {
		t.Fatal(err)
	}
	if report.MatchCount != 3 {
		t.Errorf("matches: got %d, want 3", report.MatchCount)
	}
	if report.Wins != 2 {
		t.Errorf("wins: got %d, want 2", report.Wins)
	}
	if report.Losses != 1 {
		t.Errorf("losses: got %d, want 1", report.Losses)
	}
	if report.Draws != 0 {
		t.Errorf("draws: got %d, want 0", report.Draws)
	}
	if report.TotalKills != 48 {
		t.Errorf("kills: got %d, want 48", report.TotalKills)
	}
	if report.TotalDeaths != 36 {
		t.Errorf("deaths: got %d, want 36", report.TotalDeaths)
	}
	if report.TotalAssists != 18 {
		t.Errorf("assists: got %d, want 18", report.TotalAssists)
	}
	if report.AverageKills != 16 {
		t.Errorf("average kills: got %f, want 16", report.AverageKills)
	}
	if report.AverageDeaths != 12 {
		t.Errorf("average deaths: got %f, want 12", report.AverageDeaths)
	}
	if report.AverageAssists != 6 {
		t.Errorf("average assists: got %f, want 6", report.AverageAssists)
	}
	if math.Abs(report.WinRate-200.0/3.0) > 1e-9 {
		t.Errorf("win rate: got %f, want about 66.67", report.WinRate)
	}
}

func TestAnalyzeCases(t *testing.T) {
	tests := []struct {
		name    string
		matches []Match
		want    Report
	}{
		{
			name:    "one draw with zero stats",
			matches: []Match{{Result: "draw"}},
			want:    Report{MatchCount: 1, Draws: 1},
		},
		{
			name:    "draw counts toward win rate",
			matches: []Match{{Result: "win"}, {Result: "draw"}},
			want:    Report{MatchCount: 2, Wins: 1, Draws: 1, WinRate: 50},
		},
		{
			name:    "fractional average and zero deaths",
			matches: []Match{{Result: "win", Kills: 1}, {Result: "loss"}},
			want:    Report{MatchCount: 2, Wins: 1, Losses: 1, TotalKills: 1, WinRate: 50, AverageKills: 0.5},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := AnalyzeMatches(test.matches)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestAnalyzeInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		matches []Match
	}{
		{"nil", nil},
		{"empty", []Match{}},
		{"missing result", []Match{{}}},
		{"unknown result", []Match{{Result: "WIN"}}},
		{"negative kills", []Match{{Result: "win", Kills: -1}}},
		{"negative deaths", []Match{{Result: "win", Deaths: -1}}},
		{"negative assists", []Match{{Result: "win", Assists: -1}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := AnalyzeMatches(test.matches)
			if err == nil {
				t.Error("expected invalid input error")
			}
		})
	}
}

func TestAnalyzeOverflow(t *testing.T) {
	tests := []struct {
		name    string
		matches []Match
	}{
		{"kills", []Match{{Result: "win", Kills: math.MaxInt}, {Result: "win", Kills: 1}}},
		{"deaths", []Match{{Result: "loss", Deaths: math.MaxInt}, {Result: "loss", Deaths: 1}}},
		{"assists", []Match{{Result: "draw", Assists: math.MaxInt}, {Result: "draw", Assists: 1}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := AnalyzeMatches(test.matches)
			if err == nil {
				t.Error("expected overflow error")
			}
		})
	}
}
