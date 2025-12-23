package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"backend/internal/client"
	"backend/internal/room"
	"backend/internal/storage"
	"backend/internal/utils"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// allow everything in dev
			if os.Getenv("ENV") != "production" {
				return true
			}
			// only allow prod url in prod
			origin := r.Header.Get("Origin")
			allowedOrigin := strings.TrimSuffix(os.Getenv("PROD_APP_URL"), "/")
			return origin == allowedOrigin
		},
	}
)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("roomCode")
	docIdString := r.URL.Query().Get("docId")

	if roomCode == "" || docIdString == "" {
		http.Error(w, "document id and room code are required", http.StatusBadRequest)
		return
	}

	roomCode = strings.TrimSpace(roomCode)
	docId, err := strconv.Atoi(docIdString)
	if err != nil {
		http.Error(w, "document id must be a number", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading to websocket: %v", err)
		return
	}

	rm := room.GetOrCreateRoom(roomCode)
	userId := utils.GenerateUserID()

	currentContent, err := storage.GetDocumentContent(roomCode, docId)
	if err != nil {
		log.Printf("error getting doc %d from storage: %v", docId, err)
	}

	c := client.CreateClient(userId, roomCode, docId, conn)

	// add client to the room
	room.AddClient(rm, c)

	count := room.GetClientCount(rm, docId)

	if err := client.SendInitialState(c, currentContent, count); err != nil {
		log.Printf("error sending initial state: %v", err)
	}

	// read and write continuously
	go client.WriteClient(c)
	go client.ReadClient(c, rm)

	room.BroadcastClientCount(rm, docId)
}
