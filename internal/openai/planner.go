package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/TheAlonso95/ai-dev-agent/internal/tasks"
	. "github.com/TheAlonso95/ai-dev-agent/internal/tasks"
)

func AskForTasks(prompt, apiKey string) ([]tasks.Task, error) {
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

	reqData := ChatRequest{
		Model: "o4-mini-2025-04-16",
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a project planner. Return an array of JSON tasks like: [{\"title\": \"Task 1\", \"body\": \"Do this...\"}, ...]"},
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
