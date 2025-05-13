package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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

// InitializeRepoWithReadme creates the first commit for an empty repository with a README file
func InitializeRepoWithReadme(repo string, file File, token string) error {
	owner := os.Getenv("GITHUB_USERNAME")

	// For empty repositories, we need to create a commit without a parent

	// 1. Create a blob for the README file
	blobSHA, err := createBlob(owner, repo, file, token)
	if err != nil {
		fmt.Printf("Error creating blob: %v\n", err)
		return err
	}

	// 2. Create a tree without a base tree
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees", owner, repo)
	body := map[string]interface{}{
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
		fmt.Printf("Error creating initial tree: %v\n", err)
		return err
	}

	var treeResult struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(resp, &treeResult); err != nil {
		fmt.Printf("Error parsing tree response: %v\n", err)
		return err
	}

	// 3. Create a commit without a parent
	commitURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits", owner, repo)
	commitBody := map[string]interface{}{
		"message": fmt.Sprintf("Initial commit: Add %s", file.Path),
		"tree":    treeResult.SHA,
		// No parents for initial commit
	}
	commitData, _ := json.Marshal(commitBody)
	commitResp, err := doPost(commitURL, commitData, token)
	if err != nil {
		fmt.Printf("Error creating initial commit: %v\n", err)
		return err
	}

	var commitResult struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(commitResp, &commitResult); err != nil {
		fmt.Printf("Error parsing commit response: %v\n", err)
		return err
	}

	// 4. Create the main branch reference
	refURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", owner, repo)
	refBody := map[string]interface{}{
		"ref": "refs/heads/main",
		"sha": commitResult.SHA,
	}
	refData, _ := json.Marshal(refBody)
	_, err = doPost(refURL, refData, token)
	if err != nil {
		fmt.Printf("Error creating main branch reference: %v\n", err)
		return err
	}

	return nil
}

func CreateBranchAndCommit(repo string, files []File, token string) error {
	owner := os.Getenv("GITHUB_USERNAME")

	// Check if the repository is empty
	_, _, err := getBaseCommitAndTree(owner, repo, token)
	if err != nil {
		// If we get an error that the repo is empty and we have at least one file, try initializing
		if len(files) > 0 && strings.Contains(err.Error(), "Git Repository is empty") {
			fmt.Println("Empty repository detected. Using initialization process...")
			return InitializeRepoWithReadme(repo, files[0], token)
		}
		fmt.Printf("Error getting base commit and tree: %v\n", err)
		return err
	}

	// If repository already has commits, proceed with normal process
	for _, file := range files {
		blobSHA, err := createBlob(owner, repo, file, token)
		if err != nil {
			fmt.Printf("Error creating blob: %v\n", err)
			return err
		}

		baseCommitSHA, baseTreeSHA, err := getBaseCommitAndTree(owner, repo, token)
		if err != nil {
			fmt.Printf("Error getting base commit and tree: %v\n", err)
			return err
		}

		treeSHA, err := createTree(owner, repo, file, blobSHA, baseTreeSHA, token)
		if err != nil {
			fmt.Printf("Error creating tree: %v\n", err)
			return err
		}

		commitSHA, err := createCommit(owner, repo, file, treeSHA, baseCommitSHA, token)
		if err != nil {
			fmt.Printf("Error creating commit: %v\n", err)
			return err
		}

		err = updateRef(owner, repo, commitSHA, token)
		if err != nil {
			fmt.Printf("Error updating reference: %v\n", err)
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
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/main", owner, repo)
	resp, err := doGet(url, token)
	if err != nil {
		return "", "", err
	}
	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	err = json.Unmarshal(resp, &ref)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal ref response: %w", err)
	}
	if ref.Object.SHA == "" {
		return "", "", fmt.Errorf("could not get reference SHA, response: %s", string(resp))
	}

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
	err = json.Unmarshal(resp, &commit)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal commit response: %w", err)
	}
	if commit.SHA == "" || commit.Tree.SHA == "" {
		return "", "", fmt.Errorf("could not get commit or tree SHA, response: %s", string(resp))
	}
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
	body := map[string]interface{}{
		"sha":   commitSHA,
		"force": true,
	}
	data, _ := json.Marshal(body)
	resp, err := doPatch(url, data, token)
	if err != nil {
		return err
	}

	// Check if the response indicates an error
	if len(resp) > 0 {
		// Try to determine if there was an error from the response
		var errorResp map[string]interface{}
		if jsonErr := json.Unmarshal(resp, &errorResp); jsonErr == nil {
			if message, ok := errorResp["message"].(string); ok && message != "" {
				return fmt.Errorf("GitHub API error: %s", message)
			}
		}
	}

	return nil
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return body, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return body, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return body, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}
