package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"

	"backend/internal/room"
	"backend/internal/storage"
)

func HandleGetDocuments(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("roomCode")

	if roomCode == "" {
		http.Error(w, "room code is required", http.StatusBadRequest)
		return
	}

	docs, err := storage.GetDocuments(roomCode)

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"documents": docs,
	}

	if err != nil {
		response["documents"] = []interface{}{}
	}

	json.NewEncoder(w).Encode(response)
}

func HandleCreateDocument(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("roomCode")

	if roomCode == "" {
		http.Error(w, "room code is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		req.Title = "Untitled" // default doc name
	}

	req.Title = html.EscapeString(req.Title)

	docId, err := storage.CreateDocument(roomCode, req.Title, fmt.Sprintf("# %s\n\n", req.Title))
	if err != nil {
		log.Printf("error creating document: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rm := room.GetRoom(roomCode)
	if rm != nil {
		room.BroadcastDocumentListUpdate(rm)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    docId,
		"title": req.Title,
	})
}
