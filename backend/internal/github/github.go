package github

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	GITHUB_API_BASE    = "https://api.github.com"
	GITHUB_API_VERSION = "2022-11-28"
)

func CreateRepository(accessToken, repoName string) (string, error) {
	client := &http.Client{}

	reqBody := map[string]interface{}{
		"name":        repoName,
		"description": "Study Hub - PDF storage for collaborative documents",
		"private":     false,
	}

	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/user/repos", GITHUB_API_BASE), bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", GITHUB_API_VERSION)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("GitHub API error: %d - %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("failed to create repository: status %d", resp.StatusCode)
	}

	var respData struct {
		FullName string `json:"full_name"`
		HtmlUrl  string `json:"html_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return respData.HtmlUrl, nil
}

func GetOrCreateRepository(accessToken, repoName string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/repos", GITHUB_API_BASE), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", GITHUB_API_VERSION)

	q := req.URL.Query()
	q.Add("per_page", "100")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to list repositories: %w", err)
	}
	defer resp.Body.Close()

	var repos []struct {
		Name    string `json:"name"`
		HtmlUrl string `json:"html_url"`
	}

	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&repos)
		for _, repo := range repos {
			if repo.Name == repoName {
				return repo.HtmlUrl, nil
			}
		}
	}

	return CreateRepository(accessToken, repoName)
}

func UploadFile(accessToken, owner, repo, filePath string, fileContent []byte) (string, error) {
	client := &http.Client{}

	encodedContent := base64.StdEncoding.EncodeToString(fileContent)

	reqBody := map[string]interface{}{
		"message": fmt.Sprintf("Add PDF: %s", filePath),
		"content": encodedContent,
	}

	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/repos/%s/%s/contents/pdfs/%s", GITHUB_API_BASE, owner, repo, filePath)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", GITHUB_API_VERSION)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("GitHub API error: %d - %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("failed to upload file: status %d", resp.StatusCode)
	}

	var respData struct {
		Content struct {
			HtmlUrl string `json:"html_url"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	rawUrl := convertToRawUrl(respData.Content.HtmlUrl)

	return rawUrl, nil
}

func DeleteFile(accessToken, owner, repo, filePath, sha string) error {
	client := &http.Client{}

	reqBody := map[string]interface{}{
		"message": fmt.Sprintf("Delete PDF: %s", filePath),
		"sha":     sha,
	}

	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/repos/%s/%s/contents/pdfs/%s", GITHUB_API_BASE, owner, repo, filePath)

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", GITHUB_API_VERSION)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("GitHub API error: %d - %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("failed to delete file: status %d", resp.StatusCode)
	}

	return nil
}

func GetFileInfo(accessToken, owner, repo, filePath string) (map[string]interface{}, error) {
	client := &http.Client{}

	url := fmt.Sprintf("%s/repos/%s/%s/contents/pdfs/%s", GITHUB_API_BASE, owner, repo, filePath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", GITHUB_API_VERSION)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file not found: status %d", resp.StatusCode)
	}

	var fileInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return fileInfo, nil
}

func ParseRepoFromUrl(repoUrl string) (string, string, error) {
	parts := strings.Split(strings.TrimSuffix(repoUrl, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL")
	}
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]
	return owner, repo, nil
}

func convertToRawUrl(htmlUrl string) string {
	rawUrl := strings.Replace(htmlUrl, "github.com", "raw.githubusercontent.com", 1)
	rawUrl = strings.Replace(rawUrl, "/blob/", "/", 1)
	return rawUrl
}
