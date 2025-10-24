package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larryhudson/go-todo-list-claude/internal/database"
	"github.com/larryhudson/go-todo-list-claude/internal/models"
)

func setupTestDB(t *testing.T) *database.DB {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	return db
}

func TestGetAllTodos_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	req := httptest.NewRequest("GET", "/api/todos", nil)
	w := httptest.NewRecorder()

	handler.GetAllTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todos []models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("Expected 0 todos, got %d", len(todos))
	}
}

func TestCreateTodo(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	reqBody := models.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateTodo(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var todo models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if todo.Title != "Test Todo" {
		t.Errorf("Expected title 'Test Todo', got '%s'", todo.Title)
	}

	if todo.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", todo.Description)
	}

	if todo.Completed {
		t.Error("Expected completed to be false")
	}
}

func TestCreateTodo_MissingTitle(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	reqBody := models.CreateTodoRequest{
		Description: "Test Description",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateTodo(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetTodo(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create a todo first
	created, err := repo.Create(models.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/todos/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	handler.GetTodo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todo models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if todo.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, todo.ID)
	}
}

func TestGetTodo_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	req := httptest.NewRequest("GET", "/api/todos/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	handler.GetTodo(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUpdateTodo(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create a todo first
	_, err := repo.Create(models.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	newTitle := "Updated Title"
	completed := true
	reqBody := models.UpdateTodoRequest{
		Title:     &newTitle,
		Completed: &completed,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/todos/1", bytes.NewBuffer(body))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	handler.UpdateTodo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todo models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if todo.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", todo.Title)
	}

	if !todo.Completed {
		t.Error("Expected completed to be true")
	}
}

func TestDeleteTodo(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create a todo first
	_, err := repo.Create(models.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/todos/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	handler.DeleteTodo(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify it's deleted
	todo, err := repo.GetByID(1)
	if err != nil {
		t.Fatalf("Failed to get todo: %v", err)
	}

	if todo != nil {
		t.Error("Expected todo to be deleted")
	}
}

func TestGetAllTodos_WithSearch(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create multiple todos
	_, _ = repo.Create(models.CreateTodoRequest{
		Title:       "Buy groceries",
		Description: "Milk, eggs, bread",
	})
	_, _ = repo.Create(models.CreateTodoRequest{
		Title:       "Write report",
		Description: "Q4 sales report",
	})
	_, _ = repo.Create(models.CreateTodoRequest{
		Title:       "Call customer",
		Description: "Follow up on order",
	})

	// Test search by title
	req := httptest.NewRequest("GET", "/api/todos?search=buy", nil)
	w := httptest.NewRecorder()

	handler.GetAllTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todos []models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(todos) != 1 {
		t.Errorf("Expected 1 todo, got %d", len(todos))
	}

	if todos[0].Title != "Buy groceries" {
		t.Errorf("Expected title 'Buy groceries', got '%s'", todos[0].Title)
	}
}

func TestGetAllTodos_WithSearchInDescription(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create multiple todos
	_, _ = repo.Create(models.CreateTodoRequest{
		Title:       "Todo 1",
		Description: "Contains search term",
	})
	_, _ = repo.Create(models.CreateTodoRequest{
		Title:       "Todo 2",
		Description: "Different description",
	})

	// Test search by description
	req := httptest.NewRequest("GET", "/api/todos?search=search", nil)
	w := httptest.NewRecorder()

	handler.GetAllTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todos []models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(todos) != 1 {
		t.Errorf("Expected 1 todo, got %d", len(todos))
	}
}

func TestGetAllTodos_FilterByCompleted(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create todos
	completed := true
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Todo 1"})
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Todo 2"})

	// Mark first one as completed
	_, err := repo.Update(1, models.UpdateTodoRequest{Completed: &completed})
	if err != nil {
		t.Fatalf("Failed to update todo: %v", err)
	}

	// Test filter by completed=true
	req := httptest.NewRequest("GET", "/api/todos?completed=true", nil)
	w := httptest.NewRecorder()

	handler.GetAllTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todos []models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(todos) != 1 {
		t.Errorf("Expected 1 completed todo, got %d", len(todos))
	}

	if !todos[0].Completed {
		t.Error("Expected todo to be completed")
	}
}

func TestGetAllTodos_FilterByIncomplete(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create todos
	completed := true
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Todo 1"})
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Todo 2"})

	// Mark first one as completed
	_, err := repo.Update(1, models.UpdateTodoRequest{Completed: &completed})
	if err != nil {
		t.Fatalf("Failed to update todo: %v", err)
	}

	// Test filter by completed=false
	req := httptest.NewRequest("GET", "/api/todos?completed=false", nil)
	w := httptest.NewRecorder()

	handler.GetAllTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todos []models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(todos) != 1 {
		t.Errorf("Expected 1 incomplete todo, got %d", len(todos))
	}

	if todos[0].Completed {
		t.Error("Expected todo to be incomplete")
	}
}

func TestGetAllTodos_SortByTitle(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create todos
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Zebra"})
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Apple"})
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Banana"})

	// Test sort by title ascending
	req := httptest.NewRequest("GET", "/api/todos?sortBy=title&sortOrder=asc", nil)
	w := httptest.NewRecorder()

	handler.GetAllTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todos []models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(todos) != 3 {
		t.Errorf("Expected 3 todos, got %d", len(todos))
	}

	if todos[0].Title != "Apple" {
		t.Errorf("Expected first title 'Apple', got '%s'", todos[0].Title)
	}

	if todos[1].Title != "Banana" {
		t.Errorf("Expected second title 'Banana', got '%s'", todos[1].Title)
	}

	if todos[2].Title != "Zebra" {
		t.Errorf("Expected third title 'Zebra', got '%s'", todos[2].Title)
	}
}

func TestGetAllTodos_CombinedFiltersAndSort(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := database.NewTodoRepository(db)
	handler := NewTodoHandler(repo)

	// Create todos
	completed := true
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Buy milk", Description: "grocery item"})
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Buy bread", Description: "grocery item"})
	_, _ = repo.Create(models.CreateTodoRequest{Title: "Write email", Description: "work task"})

	// Mark first two as completed
	_, err := repo.Update(1, models.UpdateTodoRequest{Completed: &completed})
	if err != nil {
		t.Fatalf("Failed to update todo: %v", err)
	}
	_, err = repo.Update(2, models.UpdateTodoRequest{Completed: &completed})
	if err != nil {
		t.Fatalf("Failed to update todo: %v", err)
	}

	// Test search + completed filter + sort
	req := httptest.NewRequest("GET", "/api/todos?search=buy&completed=true&sortBy=title&sortOrder=asc", nil)
	w := httptest.NewRecorder()

	handler.GetAllTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var todos []models.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(todos))
	}

	// Should be sorted by title
	if todos[0].Title != "Buy bread" {
		t.Errorf("Expected first title 'Buy bread', got '%s'", todos[0].Title)
	}

	if todos[1].Title != "Buy milk" {
		t.Errorf("Expected second title 'Buy milk', got '%s'", todos[1].Title)
	}
}
