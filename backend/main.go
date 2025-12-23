package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"backend/internal/ai"
	"backend/internal/auth"
	"backend/internal/handlers"
	"backend/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	PORT string = "3000"
)

type Response struct {
	Message string `json:"message"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := Response{Message: "hello there"}
	json.NewEncoder(w).Encode(response)
}

// initialize redis and postgres connections
func initializeConnections() error {
	// connect to redis
	redisHost := os.Getenv("REDIS_HOST")
	if os.Getenv("ENV") != "production" {
		redisHost = "localhost"
	}
	redisAddress := fmt.Sprintf("%s:%s", redisHost, os.Getenv("REDIS_PORT"))

	err := storage.InitializeRedis(redisAddress)
	if err != nil {
		return fmt.Errorf("could not connect to redis: %v", err)
	}

	fmt.Println("connected to redis")

	// connect to postgres

	postgrestHost := os.Getenv("POSTGRES_HOST")
	if os.Getenv("ENV") != "production" {
		postgrestHost = "localhost"
	}

	err = storage.InitializePostgres(
		postgrestHost,
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	if err != nil {
		return fmt.Errorf("could not connect to postgres: %v", err)
	}

	fmt.Println("connected to postgres")

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("could not load environment: %v", err)
	}

	auth.InitStore(os.Getenv("SESSION_SECRET"))
	ai.InitOpenAIClient(os.Getenv("OPENAI_API_KEY"))

	if err := initializeConnections(); err != nil {
		log.Fatalf("could not initialize database connections: %v", err)
		return
	}

	storage.StartBackgroundSync(2 * time.Minute)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// enable CORS for development (when frontend runs on different port)
	if os.Getenv("ENV") != "production" {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:3001"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	r.Route("/api", func(r chi.Router) {
		r.Get("/hello", helloHandler)

		r.Post("/auth/signup", auth.SignUpHandler)
		r.Post("/auth/signin", auth.SignInHandler)
		r.Get("/auth/signout", auth.SignOutHandler)
		r.Post("/auth/github", auth.GithubLoginHandler)

		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Get("/protected", auth.ProtectedRoute)

			// room endpoints
			r.Get("/rooms", handlers.HandleGetRooms)
			r.Get("/rooms/{id}", handlers.HandleGetRoom)
			r.Post("/rooms", handlers.HandleCreateRoom)

			// document endpoints
			r.Get("/documents", handlers.HandleGetDocuments)
			r.Post("/documents", handlers.HandleCreateDocument)

			// pdf endpoints
			r.Get("/pdfs", handlers.HandleGetPDFs)
			r.Post("/pdfs/upload", handlers.HandleUploadPDF)
			r.Delete("/pdfs", handlers.HandleDeletePDF)

			// ai endpoint
			r.Post("/ai", handlers.AIHandler)

			// websocket
			r.Get("/ws", handlers.HandleWebSocket)
		})
	})

	// serve static files if env is not production
	if os.Getenv("ENV") != "production" {
		staticDirectory := "../frontend/out"
		fs := http.FileServer(http.Dir(staticDirectory))
		r.Handle("/*", fs)
	}

	fmt.Printf("Server starting on port %s\n", PORT)
	if err := http.ListenAndServe(":"+PORT, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
