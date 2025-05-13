package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTP utility functions for GitHub API requests

func doGet(url, token string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
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
	
	body, err := io.ReadAll(resp.Body)
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
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode >= 300 {
		return body, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	return body, nil
}

// ParseErrorFromResponse attempts to extract an error message from a JSON response
func ParseErrorFromResponse(resp []byte) error {
	if len(resp) == 0 {
		return nil
	}
	
	var errorResp map[string]interface{}
	if err := json.Unmarshal(resp, &errorResp); err != nil {
		return nil
	}
	
	if message, ok := errorResp["message"].(string); ok && message != "" {
		return fmt.Errorf("GitHub API error: %s", message)
	}
	
	return nil
}