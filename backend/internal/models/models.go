package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	DocId    int
	RoomCode string
	Conn     *websocket.Conn
	SendChan chan []byte
	Mu       sync.Mutex
}

type Message struct {
	Type       string     `json:"type"`
	Content    string     `json:"content,omitempty"`
	UserID     string     `json:"userId,omitempty"`
	Operation  *Operation `json:"operation,omitempty"`
	Count      int        `json:"count,omitempty"`
	RoomCode   string     `json:"roomCode,omitempty"`
	DocumentId int        `json:"documentId,omitempty"`
	Message    string     `json:"message,omitempty"`
}

type Operation struct {
	Type     string `json:"type"`
	Position int    `json:"position"`
	Text     string `json:"text,omitempty"`
	Length   int    `json:"length,omitempty"`
}

// represents room internally
type Room struct {
	Code         string
	Clients      map[string]*Client
	LastActivity time.Time
	Mu           sync.RWMutex
}

type RoomInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	ActiveUsers int    `json:"activeUsers"`
	LastUpdated string `json:"lastUpdated"`
}

type Document struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type AIRequest struct {
	Prompt         string `json:"prompt"`
	DocId          int    `json:"documentId"`
	RoomCode       string `json:"roomCode"`
	CursorPosition int    `json:"cursorPosition"`
}
