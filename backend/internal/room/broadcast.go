package room

import (
	"encoding/json"

	"backend/internal/models"
)

func BroadcastClientCount(room *models.Room, docId int) {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	count := GetClientCount(room, docId)

	msg := models.Message{
		Type:  "clientCount",
		Count: count,
	}

	data, _ := json.Marshal(msg)

	for _, client := range room.Clients {
		if client.DocId == docId {
			select {
			case client.SendChan <- data:
				// sent message
			default:
				// channel full, skip client
			}
		}
	}

}

func BroadcastToOthers(room *models.Room, authorId string, docId int, data []byte) {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	for id, client := range room.Clients {
		// send update to everyone except the author
		if id != authorId && client.DocId == docId {
			select {
			case client.SendChan <- data:
				// message sent
			default:
				// channel full, skip client
			}
		}
	}
}

func BroadcastToEveryone(room *models.Room, docId int, data []byte) {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	for _, client := range room.Clients {
		// send update to everyone
		select {
		case client.SendChan <- data:
			// message sent
		default:
			// channel full, skip client
		}
	}
}

func BroadcastDocumentListUpdate(rm *models.Room) {
	msg := models.Message{
		Type: "documentListUpdate",
	}

	data, _ := json.Marshal(msg)

	rm.Mu.RLock()
	defer rm.Mu.RUnlock()

	for _, client := range rm.Clients {
		select {
		case client.SendChan <- data:
			// sent
		default:
			// skip, channel full
		}
	}
}
