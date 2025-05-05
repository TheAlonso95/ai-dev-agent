package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/TheAlonso95/ai-dev-agent/internal/tasks"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func AskForTasks(prompt, apiKey string) ([]Task, error) {
	systemPrompt := "You are an expert software project planner.\n" +
		"Given a project idea and tech stack, generate a list of development tasks formatted as JSON.\n" +
		"Each task must include:\n" +
		"- title: a short task summary\n" +
		"- body: a detailed description using markdown\n" +
		"- acceptance_criteria: a list of conditions that must be true for the task to be considered complete\n" +
		"- labels: array of relevant labels (e.g., setup, backend, frontend, auth, db)\n\n" +
		"Generate 5–10 high-quality tasks that follow best practices. Keep tasks atomic and suitable for GitHub Issues." +
		"Return ONLY a JSON array of tasks using this format:\n" +
		"[{\"title\": \"Task\", \"body\": \"...\", \"acceptance_criteria\": [...], \"labels\": [\"...\"]}]"

	reqData := ChatRequest{
		Model: "o4-mini-2025-04-16",
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	}
	jsonData, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result ChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("OpenAI error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned by OpenAI")
	}

	// Parse the JSON from the returned content
	var taskList []Task
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &taskList); err != nil {
		return nil, fmt.Errorf("failed to parse task JSON: %w", err)
	}

	return taskList, nil
}
