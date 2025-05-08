package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/TheAlonso95/ai-dev-agent/internal/tasks"
)

type Repo struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
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
		body, _ := ioutil.ReadAll(resp.Body)
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
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("Failed to decode issue: %w", err)
	}

	return &issue, nil
}

func CreateBranchAndCommit(repo string, files []File, token string) error {
	owner := os.Getenv("GITHUB_USERNAME")
	for _, file := range files {
		// Step 1: Create blob
		blobSHA, err := createBlob(owner, repo, file, token)
		if err != nil {
			return err
		}

		// Step 2: Get current commit and tree
		baseCommitSHA, baseTreeSHA, err := getBaseCommitAndTree(owner, repo, token)
		if err != nil {
			return err
		}

		// Step 3: Create tree
		treeSHA, err := createTree(owner, repo, file, blobSHA, baseTreeSHA, token)
		if err != nil {
			return err
		}

		// Step 4: Create commit
		commitSHA, err := createCommit(owner, repo, file, treeSHA, baseCommitSHA, token)
		if err != nil {
			return err
		}

		// Step 5: Update the reference to point to new commit
		err = updateRef(owner, repo, commitSHA, token)
		if err != nil {
			return err
		}
	}
	return nil
}

func createBlob(owner, repo string, file File, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/blobs", owner, repo)
	body := map[string]string{
		"content":  file.Content,
		"encoding": "utf-8",
	}
	data, _ := json.Marshal(body)
	resp, err := doPost(url, data, token)
	if err != nil {
		return "", err
	}
	var result struct {
		SHA string `json:"sha"`
	}
	json.Unmarshal(resp, &result)
	return result.SHA, nil
}

func getBaseCommitAndTree(owner, repo, token string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/heads/main", owner, repo)
	resp, err := doGet(url, token)
	if err != nil {
		return "", "", err
	}
	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	json.Unmarshal(resp, &ref)

	commitURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits/%s", owner, repo, ref.Object.SHA)
	resp, err = doGet(commitURL, token)
	if err != nil {
		return "", "", err
	}
	var commit struct {
		SHA  string `json:"sha"`
		Tree struct {
			SHA string `json:"sha"`
		} `json:"tree"`
	}
	json.Unmarshal(resp, &commit)
	return commit.SHA, commit.Tree.SHA, nil
}

func createTree(owner, repo string, file File, blobSHA, baseTreeSHA, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees", owner, repo)
	body := map[string]interface{}{
		"base_tree": baseTreeSHA,
		"tree": []map[string]string{
			{
				"path": file.Path,
				"mode": "100644",
				"type": "blob",
				"sha":  blobSHA,
			},
		},
	}
	data, _ := json.Marshal(body)
	resp, err := doPost(url, data, token)
	if err != nil {
		return "", err
	}
	var result struct {
		SHA string `json:"sha"`
	}
	json.Unmarshal(resp, &result)
	return result.SHA, nil
}

func createCommit(owner, repo string, file File, treeSHA, parentSHA, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits", owner, repo)
	body := map[string]interface{}{
		"message": fmt.Sprintf("docs: add %s", file.Path),
		"tree":    treeSHA,
		"parents": []string{parentSHA},
	}
	data, _ := json.Marshal(body)
	resp, err := doPost(url, data, token)
	if err != nil {
		return "", err
	}
	var result struct {
		SHA string `json:"sha"`
	}
	json.Unmarshal(resp, &result)
	return result.SHA, nil
}

func updateRef(owner, repo, commitSHA, token string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/main", owner, repo)
	body := map[string]string{
		"sha":   commitSHA,
		"force": "true",
	}
	data, _ := json.Marshal(body)
	_, err := doPatch(url, data, token)
	return err
}

func doGet(url, token string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func doPost(url string, data []byte, token string) ([]byte, error) {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func doPatch(url string, data []byte, token string) ([]byte, error) {
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
