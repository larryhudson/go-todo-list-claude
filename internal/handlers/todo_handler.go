package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/larryhudson/go-todo-list-claude/internal/database"
	"github.com/larryhudson/go-todo-list-claude/internal/models"
)

// TodoHandler handles HTTP requests for todos
type TodoHandler struct {
	repo *database.TodoRepository
}

// NewTodoHandler creates a new TodoHandler
func NewTodoHandler(repo *database.TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// At this point headers are already sent, so we can only log the error
		// In a production app, you'd want to use a proper logger here
		return
	}
}

// writeError writes an error JSON response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// GetAllTodos handles GET /api/todos
// @Summary Get all todos
// @Description Get all todo items with optional filtering and search
// @Tags todos
// @Produce json
// @Param search query string false "Search in title and description"
// @Param completed query boolean false "Filter by completion status"
// @Param sortBy query string false "Sort by field (createdAt, updatedAt, title)"
// @Param sortOrder query string false "Sort order (asc, desc)"
// @Success 200 {array} models.Todo
// @Failure 500 {object} ErrorResponse
// @Router /api/todos [get]
func (h *TodoHandler) GetAllTodos(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	search := r.URL.Query().Get("search")
	completedStr := r.URL.Query().Get("completed")
	sortBy := r.URL.Query().Get("sortBy")
	sortOrder := r.URL.Query().Get("sortOrder")

	// Build filter options
	opts := database.FilterOptions{
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	// Parse completed filter if provided
	if completedStr != "" {
		completed := completedStr == "true"
		opts.Completed = &completed
	}

	// If no filters provided, use GetAll for backward compatibility
	var todos []models.Todo
	var err error

	if search == "" && opts.Completed == nil && sortBy == "" {
		todos, err = h.repo.GetAll()
	} else {
		todos, err = h.repo.Search(opts)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if todos == nil {
		todos = []models.Todo{}
	}

	writeJSON(w, http.StatusOK, todos)
}

// GetTodo handles GET /api/todos/{id}
// @Summary Get a todo by ID
// @Description Get a single todo item by ID
// @Tags todos
// @Produce json
// @Param id path int true "Todo ID"
// @Success 200 {object} models.Todo
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/todos/{id} [get]
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	todo, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if todo == nil {
		writeError(w, http.StatusNotFound, "Todo not found")
		return
	}

	writeJSON(w, http.StatusOK, todo)
}

// CreateTodo handles POST /api/todos
// @Summary Create a new todo
// @Description Create a new todo item
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body models.CreateTodoRequest true "Todo to create"
// @Success 201 {object} models.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/todos [post]
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "Title is required")
		return
	}

	todo, err := h.repo.Create(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, todo)
}

// UpdateTodo handles PATCH /api/todos/{id}
// @Summary Update a todo
// @Description Update an existing todo item
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param todo body models.UpdateTodoRequest true "Todo updates"
// @Success 200 {object} models.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/todos/{id} [patch]
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req models.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	todo, err := h.repo.Update(id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if todo == nil {
		writeError(w, http.StatusNotFound, "Todo not found")
		return
	}

	writeJSON(w, http.StatusOK, todo)
}

// DeleteTodo handles DELETE /api/todos/{id}
// @Summary Delete a todo
// @Description Delete a todo item by ID
// @Tags todos
// @Param id path int true "Todo ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/todos/{id} [delete]
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Todo not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
