package main

import (
	"fmt"
	"os"
	"strings"
)

type JevResponse struct {
	Model   string     `json:"model"`
	Answers JevAnswers `json:"answers"`
	Usage   JevUsage   `json:"usage"`
}

type JevAnswers struct {
	WorthCapturing NoulAnswer   `json:"worth_capturing"`
	Kind           ChoiceAnswer `json:"kind"`
}

type NoulAnswer struct {
	Type string  `json:"type"`
	Noul float64 `json:"noul"`
}

type ChoiceAnswer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

type JevUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type KindScore struct {
	Kind  string
	Score float64
}

type Judgement struct {
	Write      bool
	WriteScore float64
	Kinds      []KindScore
}

func judge(text string) Judgement {
	return Judgement{
		Write:      true,
		WriteScore: 0.87,
		Kinds: []KindScore{
			{Kind: "self_observation", Score: 0.61},
			{Kind: "idea", Score: 0.24},
			{Kind: "learning", Score: 0.10},
			{Kind: "question", Score: 0.05},
		},
	}
}

func printJudgement(j Judgement) {
	fmt.Printf("write: %t\n", j.Write)
	fmt.Printf("score: %.2f\n", j.WriteScore)

	fmt.Println()
	fmt.Println("kinds:")

	for _, k := range j.Kinds {
		fmt.Printf("  %-18s %.2f\n", k.Kind, k.Score)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: jotworthy <text>")
		os.Exit(2)
	}

	text := strings.Join(os.Args[1:], " ")

	result := judge(text)

	printJudgement(result)
}
