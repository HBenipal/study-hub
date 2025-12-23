package room

import (
	"log"
	"sync"
	"time"

	"backend/internal/models"
)

var (
	rooms      = make(map[string]*models.Room)
	roomsMutex sync.RWMutex
)

func GetOrCreateRoom(roomCode string) *models.Room {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	room, exists := rooms[roomCode]
	if !exists {
		room = &models.Room{
			Code:         roomCode,
			Clients:      make(map[string]*models.Client),
			LastActivity: time.Now(),
		}

		rooms[roomCode] = room
		log.Printf("Created new room: %s", roomCode)
	}

	return room
}

func GetRoom(roomCode string) *models.Room {
	roomsMutex.RLock()
	defer roomsMutex.RUnlock()

	return rooms[roomCode]
}

func AddClient(room *models.Room, client *models.Client) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	room.Clients[client.ID] = client
	room.LastActivity = time.Now()
}

func RemoveClient(room *models.Room, client *models.Client) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	delete(room.Clients, client.ID)
	room.LastActivity = time.Now()
}

func GetClientCount(room *models.Room, docId int) int {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	count := 0
	for _, client := range room.Clients {
		if client.DocId == docId {
			count++
		}
	}

	return count
}

func GetActiveUsers(room *models.Room) int {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	return len(room.Clients)
}
