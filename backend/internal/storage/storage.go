package storage

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ctx         = context.Background()
	redisClient *redis.Client
	db          *sql.DB
)

type RoomData struct {
	Code      string
	Name      string
	UpdatedAt time.Time
}

// Initializes redis with given address
func InitializeRedis(address string) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr: address,
		DB:   0,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("could not ping redis: %w", err)
	}

	log.Printf("connected to redis")
	return nil
}

// Initializes postgres with given params
func InitializePostgres(host, port, user, password, dbname string) error {
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		dbname,
	)

	var err error

	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("could not connect to postgres: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("could not ping postgres: %w", err)
	}

	log.Printf("connected to postgres")
	return nil
}

func CreateDocument(roomCode, title, content string) (int, error) {
	var docId int
	err := db.QueryRow(
		`INSERT INTO documents (title, content, room_code) VALUES ($1, $2, $3) RETURNING id`,
		title,
		content,
		roomCode,
	).Scan(&docId)

	if err != nil {
		return -1, fmt.Errorf("error inserting document: %w", err)
	}

	return docId, nil
}

func GetDocuments(roomCode string) ([]models.Document, error) {
	rows, err := db.Query("SELECT id, title FROM documents WHERE room_code = $1 ORDER BY created_at ASC", roomCode)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var docs []models.Document

	for rows.Next() {
		var doc models.Document

		if err := rows.Scan(&doc.ID, &doc.Title); err != nil {
			log.Println("error reading document")
			continue
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// tries to get contents of a document from redis
// if not present in postgres, contents are fetched from postgres and added to redis
func GetDocumentContent(roomCode string, documentId int) (string, error) {
	docKey := fmt.Sprintf("doc:%s:%d:content", roomCode, documentId)
	currentContent, err := redisClient.Get(ctx, docKey).Result()

	if err != nil {
		// retrieve document from postgres
		var content string
		err = db.QueryRow(`SELECT content FROM documents WHERE id = $1 AND room_code = $2`, documentId, roomCode).Scan(&content)

		if err == nil {
			// document found in postgres
			currentContent = content
			redisClient.Set(ctx, docKey, currentContent, 1*time.Hour)
			return currentContent, nil
		}

		// document does not exist in postgres and redis, create in redis
		// it will eventually sync with postgres

		currentContent = content
		redisClient.Set(ctx, docKey, currentContent, 1*time.Hour)
		return currentContent, nil
	}

	// return content from redis
	return currentContent, nil
}

func UpdateDocumentContent(roomCode string, content string, documentId int) error {
	docKey := fmt.Sprintf("doc:%s:%d:content", roomCode, documentId)
	err := redisClient.Set(ctx, docKey, content, 1*time.Hour).Err()

	if err != nil {
		return fmt.Errorf("could not set document content: %w", err)
	}

	return nil
}

func CreateRoom(code, name string, isPublic bool) error {
	_, err := db.Exec("INSERT INTO rooms (code, name, public) VALUES ($1, $2, $3)", code, name, isPublic)
	return err
}

func GetRooms(limit, offset int) ([]RoomData, bool, error) {
	// get one more than the limit
	rows, err := db.Query("SELECT code, name, updated_at FROM rooms WHERE public = TRUE ORDER BY updated_at DESC LIMIT $1 OFFSET $2", limit+1, offset)
	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var rooms []RoomData

	for rows.Next() {
		var r RoomData
		if err := rows.Scan(&r.Code, &r.Name, &r.UpdatedAt); err != nil {
			// skip if cannot read row
			continue
		}

		rooms = append(rooms, r)
	}

	rowsCount := len(rooms)
	hasMoreData := rowsCount > limit

	if rowsCount > limit {
		// throw out the extra row, it has fulfilled its purpose
		rooms = rooms[:limit]
	}

	return rooms, hasMoreData, nil
}

func GetRoom(roomCode string) (string, bool, error) {
	var roomName string
	var isPublic bool
	err := db.QueryRow("SELECT name, public FROM rooms WHERE code = $1", roomCode).Scan(&roomName, &isPublic)
	if err != nil {
		return "", false, err
	}

	return roomName, isPublic, nil
}

func CreateUser(username, password_hash string) error {
	_, err := db.Query("INSERT INTO users (username, hashed_password) VALUES ($1, $2)", username, password_hash)
	return err
}

func GetUser(username string) (string, string, error) {
	var dbUsername string
	var dbPasswordHash string
	err := db.QueryRow("SELECT username, hashed_password FROM users WHERE username = $1", username).Scan(&dbUsername, &dbPasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", nil
		}

		log.Printf("db error: %v", err)
		return "", "", err
	}

	return dbUsername, dbPasswordHash, nil
}

func CreateOrUpdateGitHubUser(githubUsername, accessToken string) error {
	_, err := db.Exec(
		`INSERT INTO users (username, github_username, github_access_token, auth_method) 
		 VALUES ($1, $2, $3, 'github')
		 ON CONFLICT (github_username) DO UPDATE 
		 SET github_access_token = $3`,
		githubUsername,
		githubUsername,
		accessToken,
	)
	return err
}

func GetGitHubUser(githubUsername string) (string, string, error) {
	var username string
	var token string
	err := db.QueryRow(
		"SELECT github_username, github_access_token FROM users WHERE github_username = $1",
		githubUsername,
	).Scan(&username, &token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", nil
		}
		log.Printf("db error: %v", err)
		return "", "", err
	}
	return username, token, nil
}

func CreatePDF(roomCode, filename, githubUrl, uploadedBy string) (int, error) {
	var pdfId int
	err := db.QueryRow(
		`INSERT INTO pdfs (filename, room_code, github_url, uploaded_by) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id`,
		filename,
		roomCode,
		githubUrl,
		uploadedBy,
	).Scan(&pdfId)

	if err != nil {
		return -1, fmt.Errorf("error inserting pdf: %w", err)
	}

	return pdfId, nil
}

func GetPDFs(roomCode string) ([]map[string]interface{}, error) {
	rows, err := db.Query(
		"SELECT id, filename, github_url, uploaded_by, created_at FROM pdfs WHERE room_code = $1 ORDER BY created_at ASC",
		roomCode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pdfs []map[string]interface{}

	for rows.Next() {
		var id int
		var filename, githubUrl, uploadedBy string
		var createdAt time.Time

		if err := rows.Scan(&id, &filename, &githubUrl, &uploadedBy, &createdAt); err != nil {
			log.Println("error reading pdf")
			continue
		}

		pdfs = append(pdfs, map[string]interface{}{
			"id":          id,
			"filename":    filename,
			"github_url":  githubUrl,
			"uploaded_by": uploadedBy,
			"created_at":  createdAt,
		})
	}

	return pdfs, nil
}

func DeletePDF(pdfId int) error {
	_, err := db.Exec("DELETE FROM pdfs WHERE id = $1", pdfId)
	return err
}

func GetPDFByID(pdfId int) (string, string, error) {
	var filename, roomCode string
	err := db.QueryRow(
		"SELECT filename, room_code FROM pdfs WHERE id = $1",
		pdfId,
	).Scan(&filename, &roomCode)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("pdf not found")
		}
		return "", "", err
	}

	return filename, roomCode, nil
}

func GetPDFOwner(pdfId int) (string, error) {
	var uploadedBy string
	err := db.QueryRow(
		"SELECT uploaded_by FROM pdfs WHERE id = $1",
		pdfId,
	).Scan(&uploadedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("pdf not found")
		}
		return "", err
	}

	return uploadedBy, nil
}

func StartBackgroundSync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)

		for range ticker.C {
			keys, _ := redisClient.Keys(ctx, "doc:*:*:content").Result()
			for _, key := range keys {
				// decompose the key of form doc:{roomCode}:{docId}:content
				parts := strings.Split(key, ":")
				if len(parts) == 4 {
					roomCode := parts[1]
					docId, err := strconv.Atoi(parts[2])
					if err != nil {
						// ignore invalid doc ids
						continue
					}
					content, _ := redisClient.Get(ctx, key).Result()

					db.Exec(`UPDATE documents SET content = $1 WHERE id = $2 AND room_code = $3`, content, docId, roomCode)
				}
			}
			log.Println("synced all documents with postgres")
		}
	}()
}
