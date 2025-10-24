// Package main provides the entry point for the todo API server
// @title Todo API
// @version 1.0
// @description A simple todo list API
// @host localhost:8080
// @BasePath /
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/larryhudson/go-todo-list-claude/internal/database"
	"github.com/larryhudson/go-todo-list-claude/internal/handlers"
)

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./todos.db"
	}

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create repository and handler
	todoRepo := database.NewTodoRepository(db)
	todoHandler := handlers.NewTodoHandler(todoRepo)

	// Create router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /api/todos", todoHandler.GetAllTodos)
	mux.HandleFunc("GET /api/todos/{id}", todoHandler.GetTodo)
	mux.HandleFunc("POST /api/todos", todoHandler.CreateTodo)
	mux.HandleFunc("PATCH /api/todos/{id}", todoHandler.UpdateTodo)
	mux.HandleFunc("DELETE /api/todos/{id}", todoHandler.DeleteTodo)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing health check response: %v", err)
		}
	})

	// Wrap with CORS middleware
	handler := corsMiddleware(mux)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server with timeouts for security
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
