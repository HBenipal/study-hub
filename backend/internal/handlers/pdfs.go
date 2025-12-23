package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/auth"
	ghub "backend/internal/github"
	"backend/internal/storage"
)

func HandleGetPDFs(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("roomCode")

	if roomCode == "" {
		http.Error(w, "room code is required", http.StatusBadRequest)
		return
	}

	pdfs, err := storage.GetPDFs(roomCode)

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"pdfs": pdfs,
	}

	if err != nil {
		response["pdfs"] = []interface{}{}
	}

	json.NewEncoder(w).Encode(response)
}

func HandleUploadPDF(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsernameFromContext(r.Context())
	roomCode := r.URL.Query().Get("roomCode")

	if roomCode == "" {
		http.Error(w, "room code is required", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(50 * 1024 * 1024); err != nil { // 50MB max
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(handler.Filename), ".pdf") {
		http.Error(w, "only PDF files are allowed", http.StatusBadRequest)
		return
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	githubUsername, githubToken, err := storage.GetGitHubUser(username)
	if err != nil || githubToken == "" {
		http.Error(w, "GitHub account not linked. Please link your GitHub account to upload PDFs.", http.StatusForbidden)
		return
	}

	personalRepoUrl, err := getOrCreateUserPDFRepo(githubToken, githubUsername)
	if err != nil {
		log.Printf("error getting/creating user PDF repo: %v", err)
		http.Error(w, "failed to setup GitHub repository", http.StatusInternalServerError)
		return
	}

	owner, repo, err := ghub.ParseRepoFromUrl(personalRepoUrl)
	if err != nil {
		http.Error(w, "invalid repository URL", http.StatusInternalServerError)
		return
	}

	filePath := roomCode + "/" + handler.Filename
	githubUrl, err := ghub.UploadFile(githubToken, owner, repo, filePath, fileContent)
	if err != nil {
		log.Printf("error uploading to GitHub: %v", err)
		http.Error(w, "failed to upload file to GitHub", http.StatusInternalServerError)
		return
	}

	pdfId, err := storage.CreatePDF(roomCode, handler.Filename, githubUrl, username)
	if err != nil {
		log.Printf("error storing pdf metadata: %v", err)
		http.Error(w, "failed to store PDF metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         pdfId,
		"filename":   handler.Filename,
		"github_url": githubUrl,
	})
}

func HandleDeletePDF(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsernameFromContext(r.Context())

	pdfIdStr := r.URL.Query().Get("pdfId")
	if pdfIdStr == "" {
		http.Error(w, "pdf id is required", http.StatusBadRequest)
		return
	}

	pdfId, err := strconv.Atoi(pdfIdStr)
	if err != nil {
		http.Error(w, "invalid pdf id", http.StatusBadRequest)
		return
	}

	filename, roomCode, err := storage.GetPDFByID(pdfId)
	if err != nil {
		http.Error(w, "pdf not found", http.StatusNotFound)
		return
	}

	uploadedBy, err := storage.GetPDFOwner(pdfId)
	if err != nil {
		http.Error(w, "pdf not found", http.StatusNotFound)
		return
	}

	if uploadedBy != username {
		http.Error(w, "You can only delete PDFs you uploaded", http.StatusForbidden)
		return
	}

	_, githubToken, err := storage.GetGitHubUser(username)
	if err != nil || githubToken == "" {
		http.Error(w, "GitHub account not linked", http.StatusForbidden)
		return
	}

	repoUrl, err := getOrCreateUserPDFRepo(githubToken, username)
	if err != nil {
		log.Printf("error getting user PDF repo: %v", err)
		http.Error(w, "failed to get GitHub repository", http.StatusInternalServerError)
		return
	}

	owner, repo, err := ghub.ParseRepoFromUrl(repoUrl)
	if err != nil {
		http.Error(w, "invalid repository URL", http.StatusInternalServerError)
		return
	}

	filePath := roomCode + "/" + filename
	fileInfo, err := ghub.GetFileInfo(githubToken, owner, repo, filePath)
	if err != nil {
		log.Printf("error getting file info from GitHub: %v", err)
		http.Error(w, "failed to get file info from GitHub", http.StatusInternalServerError)
		return
	}

	sha, ok := fileInfo["sha"].(string)
	if !ok {
		http.Error(w, "failed to extract file SHA", http.StatusInternalServerError)
		return
	}

	if err := ghub.DeleteFile(githubToken, owner, repo, filePath, sha); err != nil {
		log.Printf("error deleting file from GitHub: %v", err)
		http.Error(w, "failed to delete file from GitHub", http.StatusInternalServerError)
		return
	}

	if err := storage.DeletePDF(pdfId); err != nil {
		log.Printf("error deleting pdf from database: %v", err)
		http.Error(w, "failed to delete PDF from database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "PDF deleted successfully",
	})
}

func getOrCreateUserPDFRepo(accessToken, githubUsername string) (string, error) {
	repoName := "study-hub-pdfs"
	repoUrl, err := ghub.GetOrCreateRepository(accessToken, repoName)
	if err != nil {
		return "", err
	}
	return repoUrl, nil
}
