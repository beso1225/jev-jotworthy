package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
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

func getJevResponse(text string) (JevResponse, error) {
	data, err := os.ReadFile("testdata/sample_response.json")
	if err != nil {
		return JevResponse{}, err
	}

	var response JevResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return JevResponse{}, err
	}

	return response, nil
}

func interpret(response JevResponse) Judgement {
	writeScore := response.Answers.WorthCapturing.Noul
	write := writeScore > 0.7

	kinds := sortKinds(response.Answers.Kind.Probabilities)

	return Judgement{
		Write:      write,
		WriteScore: writeScore,
		Kinds:      kinds,
	}
}

func sortKinds(probabilities map[string]float64) []KindScore {
	kinds := make([]KindScore, 0, len(probabilities))
	for kind, score := range probabilities {
		kinds = append(kinds, KindScore{Kind: kind, Score: score})
	}

	sort.Slice(kinds, func(i, j int) bool {
		if kinds[i].Score == kinds[j].Score {
			return kinds[i].Kind < kinds[j].Kind
		}
		return kinds[i].Score > kinds[j].Score
	})

	return kinds
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

	response, err := getJevResponse(text)
	if err != nil {
		log.Fatal(err)
	}

	result := interpret(response)

	printJudgement(result)
}
