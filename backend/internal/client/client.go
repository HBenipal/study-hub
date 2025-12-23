package client

import (
	"backend/internal/models"
	"backend/internal/storage"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"time"

	"backend/internal/room"
	"backend/internal/utils"

	"github.com/gorilla/websocket"
)

func CreateClient(userId string, roomCode string, docId int, conn *websocket.Conn) *models.Client {
	return &models.Client{
		ID:       userId,
		DocId:    docId,
		Conn:     conn,
		RoomCode: roomCode,
		SendChan: make(chan []byte, 256),
	}
}

// continuosly writes messages to given client
func WriteClient(client *models.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		// there is a message to be sent
		case message, ok := <-client.SendChan:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			client.Mu.Lock()
			err := client.Conn.WriteMessage(websocket.TextMessage, message)
			client.Mu.Unlock()

			if err != nil {
				log.Printf("could not write to client id %s\n", client.ID)
				return
			}

		// ping client to keep connection alive
		case <-ticker.C:
			client.Mu.Lock()
			err := client.Conn.WriteMessage(websocket.PingMessage, nil)
			client.Mu.Unlock()

			if err != nil {
				log.Printf("could not send ping to client id %s\n", client.ID)
				return
			}
		}
	}
}

func ReadClient(client *models.Client, rm *models.Room) {
	defer func() {
		room.RemoveClient(rm, client)
		close(client.SendChan)
		client.Conn.Close()
		log.Printf("Client %s disconnected\n", client.ID)
		room.BroadcastClientCount(rm, client.DocId)
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("could not unmarshal message: %v", err)
			continue
		}

		switch msg.Type {
		case "operation":
			handleOperation(client, rm, &msg)
		}
	}
}

func SendInitialState(client *models.Client, content string, count int) error {
	initMsg := models.Message{
		Type:    "init",
		Content: content,
		UserID:  client.ID,
		Count:   count,
	}

	data, err := json.Marshal(initMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal init message: %w", err)
	}

	client.SendChan <- data
	return nil
}

func handleOperation(client *models.Client, rm *models.Room, msg *models.Message) {
	if msg.Operation == nil {
		return
	}

	currentContent, err := storage.GetDocumentContent(client.RoomCode, client.DocId)
	if err != nil {
		log.Printf("error getting document %d\n", client.DocId)
		return
	}

	newContent := utils.ApplyOperation(currentContent, msg.Operation)

	err = storage.UpdateDocumentContent(client.RoomCode, newContent, client.DocId)
	if err != nil {
		log.Printf("could not update document %d\n", client.DocId)
	}

	msg.Operation.Text = html.EscapeString(msg.Operation.Text)

	response := models.Message{
		Type:      "operation",
		Operation: msg.Operation,
		UserID:    client.ID,
	}

	data, _ := json.Marshal(response)
	room.BroadcastToOthers(rm, client.ID, client.DocId, data)
}
