package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	. "github.com/TheAlonso95/ai-dev-agent/internal/tasks"
)

type Repo struct {
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	AutoInit bool   `json:"auto_init"`
}

type Issue struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

type File struct {
	Path    string
	Content string
}

func CreateRepo(repoName string, token string) error {
	repo := Repo{Name: repoName, Private: false, AutoInit: true}
	jsonData, _ := json.Marshal(repo)

	req, _ := http.NewRequest("POST", "https://api.github.com/user/repos", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create repo: %s", body)
	}

	return nil
}

func CreateIssue(owner, repo, token string, task Task) error {
	// Format acceptance criteria into markdown
	acSection := ""
	if len(task.AcceptanceCriteria) > 0 {
		acSection = "### Acceptance Criteria:\n"
		for _, item := range task.AcceptanceCriteria {
			acSection += fmt.Sprintf("- %s\n", item)
		}
	}

	fullBody := fmt.Sprintf("%s\n\n%s", task.Body, acSection)

	issue := map[string]interface{}{
		"title":  task.Title,
		"body":   fullBody,
		"labels": task.Labels,
	}
	jsonData, _ := json.Marshal(issue)

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("issue creation failed: %s", body)
	}

	return nil
}

func FetchIssue(owner, repo string, issueNumber int, token string) (*Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, issueNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("Failed to decode issue: %w", err)
	}

	return &issue, nil
}
