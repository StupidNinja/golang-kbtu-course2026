package main

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestDecodeSession(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Match
	}{
		{
			name: "sample session",
			input: `[
				{"result":"win","kills":18,"deaths":12,"assists":6},
				{"result":"loss","kills":10,"deaths":16,"assists":4},
				{"result":"win","kills":20,"deaths":8,"assists":8}
			]`,
			want: sampleMatches(),
		},
		{"zero win", `[{"result":"win","kills":0,"deaths":0,"assists":0}]`, []Match{{Result: "win"}}},
		{"zero loss", `[{"result":"loss","kills":0,"deaths":0,"assists":0}]`, []Match{{Result: "loss"}}},
		{"zero draw", `[{"result":"draw","kills":0,"deaths":0,"assists":0}]`, []Match{{Result: "draw"}}},
		{"trailing whitespace", "[{\"result\":\"draw\",\"kills\":0,\"deaths\":0,\"assists\":0}]\n\t ", []Match{{Result: "draw"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DecodeSession(strings.NewReader(test.input))
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(test.want) {
				t.Fatalf("got %d matches, want %d", len(got), len(test.want))
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("match %d: got %+v, want %+v", i, got[i], test.want[i])
				}
			}
		})
	}
}

func TestDecodeSessionInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty file", ""},
		{"empty array", `[]`},
		{"null root", `null`},
		{"object root", `{}`},
		{"broken JSON", `[{"result":"win"`},
		{"null match", `[null]`},
		{"missing result", `[{"kills":0,"deaths":0,"assists":0}]`},
		{"missing kills", `[{"result":"win","deaths":0,"assists":0}]`},
		{"missing deaths", `[{"result":"win","kills":0,"assists":0}]`},
		{"missing assists", `[{"result":"win","kills":0,"deaths":0}]`},
		{"null result", `[{"result":null,"kills":0,"deaths":0,"assists":0}]`},
		{"null kills", `[{"result":"win","kills":null,"deaths":0,"assists":0}]`},
		{"null deaths", `[{"result":"win","kills":0,"deaths":null,"assists":0}]`},
		{"null assists", `[{"result":"win","kills":0,"deaths":0,"assists":null}]`},
		{"unknown field", `[{"result":"win","kills":0,"deaths":0,"assists":0,"killz":10}]`},
		{"misspelled kills", `[{"result":"win","killz":10,"deaths":0,"assists":0}]`},
		{"unknown result", `[{"result":"victory","kills":0,"deaths":0,"assists":0}]`},
		{"empty result", `[{"result":"","kills":0,"deaths":0,"assists":0}]`},
		{"string counter", `[{"result":"win","kills":"1","deaths":0,"assists":0}]`},
		{"fractional counter", `[{"result":"win","kills":1.5,"deaths":0,"assists":0}]`},
		{"negative kills", `[{"result":"win","kills":-1,"deaths":0,"assists":0}]`},
		{"negative deaths", `[{"result":"win","kills":0,"deaths":-1,"assists":0}]`},
		{"negative assists", `[{"result":"win","kills":0,"deaths":0,"assists":-1}]`},
		{"outside int", fmt.Sprintf(`[{"result":"win","kills":%d,"deaths":0,"assists":0}]`, uint64(math.MaxInt)+1)},
		{"second document", `[{"result":"win","kills":0,"deaths":0,"assists":0}] []`},
		{"trailing null", `[{"result":"win","kills":0,"deaths":0,"assists":0}] null`},
		{"trailing garbage", `[{"result":"win","kills":0,"deaths":0,"assists":0}] broken`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := DecodeSession(strings.NewReader(test.input))
			if err == nil {
				t.Error("expected JSON or validation error")
			}
		})
	}
}
