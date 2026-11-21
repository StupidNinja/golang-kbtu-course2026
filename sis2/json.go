package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Указатели отличают пропущенное поле от разрешённого значения 0.
type matchInput struct {
	Result  *string `json:"result"`
	Kills   *int    `json:"kills"`
	Deaths  *int    `json:"deaths"`
	Assists *int    `json:"assists"`
}

func DecodeSession(r io.Reader) ([]Match, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var input []matchInput
	err := decoder.Decode(&input)
	if err != nil {
		return nil, fmt.Errorf("invalid session JSON: %v", err)
	}

	var extra []matchInput
	err = decoder.Decode(&extra)
	if err == nil {
		return nil, errors.New("session must contain exactly one JSON array")
	}
	if err != io.EOF {
		return nil, fmt.Errorf("invalid data after session: %v", err)
	}

	matches := make([]Match, len(input))
	for i, match := range input {
		if match.Result == nil || match.Kills == nil || match.Deaths == nil || match.Assists == nil {
			return nil, fmt.Errorf("match %d: result, kills, deaths and assists are required and cannot be null", i+1)
		}
		matches[i] = Match{
			Result:  *match.Result,
			Kills:   *match.Kills,
			Deaths:  *match.Deaths,
			Assists: *match.Assists,
		}
	}
	err = validateMatches(matches)
	if err != nil {
		return nil, err
	}
	return matches, nil
}
