package handlers

import (
	"encoding/json"
	"html"
	"net/http"

	"backend/internal/ai"
	"backend/internal/models"
)

func AIHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AIRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body must have prompt, documentId and roomCode", http.StatusBadRequest)
		return
	}

	if req.RoomCode == "" || req.Prompt == "" {
		http.Error(w, "roomCode and prompt must be non empty", http.StatusBadRequest)
		return
	}

	req.Prompt = html.EscapeString(req.Prompt)

	// since ai response can take long we can do it async and sent accepted response
	go ai.BroadcastAIResponse(&req)
	w.WriteHeader(http.StatusAccepted)
}
