package github

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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

// CreateBranchAndCommit creates a commit with the given files and pushes it to the main branch
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
