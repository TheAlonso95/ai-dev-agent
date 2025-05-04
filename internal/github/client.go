package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/TheAlonso95/ai-dev-agent/internal/tasks"
)

type Repo struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

func CreateRepo(repoName string, token string) error {
	repo := Repo{Name: repoName, Private: false}
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
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to create repo: %s", body)
	}

	return nil
}

func CreateIssue(owner, repo, token string, task Task) error {
	issue := map[string]string{
		"title": task.Title,
		"body":  task.Body,
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
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("issue creation failed: %s", body)
	}

	return nil
}
