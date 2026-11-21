package main

import (
	"errors"
	"fmt"
	"math"
)

const (
	ResultWin  = "win"
	ResultLoss = "loss"
	ResultDraw = "draw"
)

type Match struct {
	Result  string `json:"result"`
	Kills   int    `json:"kills"`
	Deaths  int    `json:"deaths"`
	Assists int    `json:"assists"`
}

type Report struct {
	MatchCount     int
	Wins           int
	Losses         int
	Draws          int
	TotalKills     int
	TotalDeaths    int
	TotalAssists   int
	WinRate        float64
	AverageKills   float64
	AverageDeaths  float64
	AverageAssists float64
}

func validateMatches(matches []Match) error {
	if len(matches) == 0 {
		return errors.New("session must contain at least one match")
	}
	for i, match := range matches {
		if match.Result != ResultWin && match.Result != ResultLoss && match.Result != ResultDraw {
			return fmt.Errorf("match %d: result must be win, loss or draw", i+1)
		}
		if match.Kills < 0 || match.Deaths < 0 || match.Assists < 0 {
			return fmt.Errorf("match %d: kills, deaths and assists cannot be negative", i+1)
		}
	}
	return nil
}

func AnalyzeMatches(matches []Match) (Report, error) {
	err := validateMatches(matches)
	if err != nil {
		return Report{}, err
	}

	report := Report{MatchCount: len(matches)}
	for i, match := range matches {
		// Проверяем сумму до сложения, чтобы int не переполнился.
		if match.Kills > math.MaxInt-report.TotalKills {
			return Report{}, fmt.Errorf("match %d: total kills exceed int range", i+1)
		}
		if match.Deaths > math.MaxInt-report.TotalDeaths {
			return Report{}, fmt.Errorf("match %d: total deaths exceed int range", i+1)
		}
		if match.Assists > math.MaxInt-report.TotalAssists {
			return Report{}, fmt.Errorf("match %d: total assists exceed int range", i+1)
		}
		report.TotalKills += match.Kills
		report.TotalDeaths += match.Deaths
		report.TotalAssists += match.Assists

		switch match.Result {
		case ResultWin:
			report.Wins++
		case ResultLoss:
			report.Losses++
		case ResultDraw:
			report.Draws++
		}
	}

	count := float64(report.MatchCount)
	report.WinRate = float64(report.Wins) / count * 100
	report.AverageKills = float64(report.TotalKills) / count
	report.AverageDeaths = float64(report.TotalDeaths) / count
	report.AverageAssists = float64(report.TotalAssists) / count
	return report, nil
}
