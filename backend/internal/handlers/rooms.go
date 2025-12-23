package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/internal/models"
	"backend/internal/room"
	"backend/internal/storage"
	"backend/internal/utils"

	"github.com/go-chi/chi/v5"
)

type RoomResponse struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Public bool   `json:"isPublic"`
}

func HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Public bool   `json:"isPublic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Name = "Untitled Hub"
		req.Public = false
	}

	if req.Name == "" {
		req.Name = "Untitled Hub"
	}

	req.Name = html.EscapeString(req.Name)

	roomCode := utils.GenerateRoomCode()

	err := storage.CreateRoom(roomCode, req.Name, req.Public)
	if err != nil {
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	// create the default document
	_, err = storage.CreateDocument(roomCode, "Untitled Document", fmt.Sprintf("# Welcome to %s!", req.Name))
	if err != nil {
		log.Printf("Error creating default document: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")

	roomResponse := RoomResponse{
		Code:   roomCode,
		Name:   req.Name,
		Public: req.Public,
	}

	json.NewEncoder(w).Encode(roomResponse)
}

func HandleGetRooms(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	if limit == "" {
		limit = "10"
	}

	if offset == "" {
		offset = "0"
	}

	limitNum, err := strconv.Atoi(limit)
	if err != nil {
		http.Error(w, "limit must be a number", http.StatusBadRequest)
		return
	}

	offsetNum, err := strconv.Atoi(offset)
	if err != nil {
		http.Error(w, "offset must be a number", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	dbRooms, hasMoreRooms, err := storage.GetRooms(limitNum, offsetNum)
	if err != nil {
		log.Printf("db error: %v", err)
		// return an empty list, no rooms found
		json.NewEncoder(w).Encode(map[string]interface{}{
			"rooms": []interface{}{},
		})
		return
	}

	roomsList := []models.RoomInfo{}
	for _, dbRoom := range dbRooms {
		activeUsers := 0
		if rm := room.GetRoom(dbRoom.Code); rm != nil {
			activeUsers = room.GetActiveUsers(rm)
		}

		lastUpdated := "recently"
		if time.Since(dbRoom.UpdatedAt) < time.Minute {
			lastUpdated = "just now"
		} else if time.Since(dbRoom.UpdatedAt) < time.Hour {
			lastUpdated = fmt.Sprintf("%d min ago", int(time.Since(dbRoom.UpdatedAt).Minutes()))
		}

		roomsList = append(roomsList, models.RoomInfo{
			Code:        dbRoom.Code,
			Name:        dbRoom.Name,
			ActiveUsers: activeUsers,
			LastUpdated: lastUpdated,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms":       roomsList,
		"hasMoreData": hasMoreRooms,
	})
}

func HandleGetRoom(w http.ResponseWriter, r *http.Request) {
	roomCode := chi.URLParam(r, "id")

	if roomCode == "" {
		http.Error(w, "room code is required", http.StatusBadRequest)
		return
	}

	roomName, isPublic, err := storage.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(RoomResponse{
		Code:   roomCode,
		Name:   roomName,
		Public: isPublic,
	})
}
