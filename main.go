package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type JevRequest struct {
	Model     string       `json:"model"`
	State     string       `json:"state"`
	Questions JevQuestions `json:"questions"`
}

type JevQuestions struct {
	WorthCapturing NoulQuestion   `json:"worth_capturing"`
	Kind           ChoiceQuestion `json:"kind"`
}

type NoulQuestion struct {
	Type         string `json:"type"`
	Instructions string `json:"instructions"`
}

type ChoiceQuestion struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

func buildJevRequest(text string) JevRequest {
	return JevRequest{
		Model: "jev-latest",
		State: text,
		Questions: JevQuestions{
			WorthCapturing: NoulQuestion{
				Type: "noul",
				Instructions: `Decide whether the user should jot this down in their personal Obsidian vault.

The vault is a low-friction personal knowledge base. Small technical facts,
tool behavior, shortcuts, debugging observations, project-specific decisions,
questions, ideas, and subjective impressions can all be worth capturing.

Do not require the note to be broadly important, novel, polished, or useful
to other people. A small detail is worth capturing if the user may reasonably
want to remember, revisit, connect, or search for it later.

The cost of keeping a low-value note is small. The cost of losing a useful
thought is higher. When uncertain, prefer capturing.`,
			},
			Kind: ChoiceQuestion{
				Type:         "choice",
				Instructions: "What kind of Obsidian Daily Note entry is this?",
				Criteria: map[string]string{
					"learning":         "Something the writer learned, understood, or wants to remember.",
					"idea":             "An idea, hypothesis, possibility, or thought worth developing.",
					"question":         "An unresolved question worth revisiting or investigating.",
					"self_observation": "An observation about the writer's own thinking, behavior, habits, or preferences.",
					"decision":         "A decision, conclusion, or choice whose reasoning may be useful later.",
					"reference":        "Information that may be useful to look up or refer back to later.",
					"other":            "Worth capturing, but none of the other categories fit well.",
				},
			},
		},
	}
}

const jevEndpoint = "https://api.typesafe.ai/v1/systemone"

func sendJevRequest(request JevRequest) (JevResponse, error) {
	apiKey := os.Getenv("JEV_API_KEY")
	if apiKey == "" {
		return JevResponse{}, fmt.Errorf("JEV_API_KEY not set")
	}

	body, err := json.Marshal(request)
	if err != nil {
		return JevResponse{}, err
	}

	httpRequest, err := http.NewRequest(
		http.MethodPost,
		jevEndpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return JevResponse{}, err
	}

	httpRequest.Header.Set("Authorization", "Bearer "+apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(httpRequest)
	if err != nil {
		return JevResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return JevResponse{}, fmt.Errorf(
			"Jev API returned %s: %s",
			resp.Status,
			string(body),
		)
	}

	var response JevResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return JevResponse{}, err
	}

	return response, nil

}

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
	write := writeScore >= 0.6

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

	request := buildJevRequest(text)

	response, err := sendJevRequest(request)
	if err != nil {
		log.Fatal(err)
	}

	result := interpret(response)

	printJudgement(result)
}
