package main

import (
	"encoding/json"
	"testing"
)

func TestParseCLIArgs(t *testing.T) {
	options, args, err := parseCLIArgs([]string{"--json", "--stdin"})
	if err != nil {
		t.Fatalf("parseCLIArgs returned an error: %v", err)
	}

	if !options.jsonOutput {
		t.Error("jsonOutput = false, want true")
	}
	if !options.stdin {
		t.Error("stdin = false, want true")
	}
	if len(args) != 0 {
		t.Fatalf("args = %#v, want no positional arguments", args)
	}
	if options.threshold != defaultThreshold {
		t.Errorf("threshold = %f, want %f", options.threshold, defaultThreshold)
	}
}

func TestParseCLIArgsThreshold(t *testing.T) {
	options, _, err := parseCLIArgs([]string{"--threshold", "0.85"})
	if err != nil {
		t.Fatalf("parseCLIArgs returned an error: %v", err)
	}

	if options.threshold != 0.85 {
		t.Errorf("threshold = %f, want 0.85", options.threshold)
	}
}

func TestParseCLIArgsRejectsInvalidThreshold(t *testing.T) {
	for _, value := range []string{"-0.01", "1.01"} {
		t.Run(value, func(t *testing.T) {
			if _, _, err := parseCLIArgs([]string{"--threshold", value}); err == nil {
				t.Fatalf("parseCLIArgs(%q) returned nil error, want an error", value)
			}
		})
	}
}

func TestJudgementJSON(t *testing.T) {
	judgement := Judgement{
		Write:      true,
		WriteScore: 0.87,
		Kinds: []KindScore{
			{Kind: "idea", Score: 0.61},
		},
	}

	data, err := json.Marshal(judgement)
	if err != nil {
		t.Fatalf("json.Marshal returned an error: %v", err)
	}

	got := string(data)
	want := `{"write":true,"write_score":0.87,"kinds":[{"kind":"idea","score":0.61}]}`
	if got != want {
		t.Errorf("JSON = %s, want %s", got, want)
	}
}

func TestSortKinds(t *testing.T) {
	probabilities := map[string]float64{
		"idea":             0.24,
		"self_observation": 0.61,
		"question":         0.05,
		"learning":         0.10,
	}

	got := sortKinds(probabilities)
	want := []string{"self_observation", "idea", "learning", "question"}

	for i, kind := range want {
		if got[i].Kind != kind {
			t.Fatalf(
				"got[%d].Kind = %q; want %q",
				i, got[i].Kind, kind,
			)
		}
	}
}

func TestSortKindsTieBreak(t *testing.T) {
	probabilities := map[string]float64{
		"idea":     0.5,
		"learning": 0.5,
	}

	got := sortKinds(probabilities)

	if got[0].Kind != "idea" {
		t.Errorf("first kind = %q, want idea", got[0].Kind)
	}
}

func TestInterpret(t *testing.T) {
	response := JevResponse{
		Answers: JevAnswers{
			WorthCapturing: NoulAnswer{
				Noul: 0.87,
			},
			Kind: ChoiceAnswer{
				Probabilities: map[string]float64{
					"idea":             0.24,
					"self_observation": 0.61,
					"learning":         0.10,
					"question":         0.05,
				},
			},
		},
	}

	got := interpret(response)

	if !got.Write {
		t.Errorf("Write = false, want true")
	}

	if got.WriteScore != 0.87 {
		t.Errorf("WriteScore = %f, want 0.87", got.WriteScore)
	}

	if got.Kinds[0].Kind != "self_observation" {
		t.Errorf(
			"first kind = %q, want self_observation",
			got.Kinds[0].Kind,
		)
	}
}

func TestInterpretThreshold(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  bool
	}{
		{"above threshold", 0.71, true},
		{"at threshold", 0.60, true},
		{"below threshold", 0.59, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := JevResponse{
				Answers: JevAnswers{
					WorthCapturing: NoulAnswer{
						Noul: tt.score,
					},
					Kind: ChoiceAnswer{
						Probabilities: map[string]float64{},
					},
				},
			}

			got := interpret(response)

			if got.Write != tt.want {
				t.Errorf(
					"score %f: Write = %v, want %v",
					tt.score,
					got.Write,
					tt.want,
				)
			}
		})
	}
}

func TestInterpretWithThreshold(t *testing.T) {
	response := JevResponse{
		Answers: JevAnswers{
			WorthCapturing: NoulAnswer{Noul: 0.65},
			Kind:           ChoiceAnswer{Probabilities: map[string]float64{}},
		},
	}

	if got := interpretWithThreshold(response, 0.60).Write; !got {
		t.Error("threshold 0.60: Write = false, want true")
	}
	if got := interpretWithThreshold(response, 0.70).Write; got {
		t.Error("threshold 0.70: Write = true, want false")
	}
}
