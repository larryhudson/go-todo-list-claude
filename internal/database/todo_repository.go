package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/larryhudson/go-todo-list-claude/internal/models"
)

// TodoRepository handles database operations for todos
type TodoRepository struct {
	db *DB
}

// NewTodoRepository creates a new TodoRepository
func NewTodoRepository(db *DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create creates a new todo
func (r *TodoRepository) Create(req models.CreateTodoRequest) (*models.Todo, error) {
	query := `
		INSERT INTO todos (title, description, completed, created_at, updated_at)
		VALUES (?, ?, 0, ?, ?)
		RETURNING id, title, description, completed, created_at, updated_at
	`

	now := time.Now()
	var todo models.Todo

	err := r.db.QueryRowContext(context.Background(), query, req.Title, req.Description, now, now).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return &todo, nil
}

// GetAll returns all todos
func (r *TodoRepository) GetAll() ([]models.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating todos: %w", err)
	}

	// Check for errors from closing rows
	if err = rows.Close(); err != nil {
		return nil, fmt.Errorf("failed to close rows: %w", err)
	}

	return todos, nil
}

// FilterOptions contains filtering and sorting options
type FilterOptions struct {
	Search    string
	Completed *bool
	SortBy    string
	SortOrder string
}

// Search searches and filters todos
func (r *TodoRepository) Search(opts FilterOptions) ([]models.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE 1=1
	`
	var args []interface{}

	// Add search filter
	if opts.Search != "" {
		query += ` AND (title LIKE ? OR description LIKE ?)`
		searchTerm := "%" + opts.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Add completion filter
	if opts.Completed != nil {
		query += ` AND completed = ?`
		args = append(args, *opts.Completed)
	}

	// Add sorting
	sortBy := "created_at"
	if opts.SortBy != "" {
		// Validate sort field to prevent SQL injection
		validFields := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"title":      true,
		}
		if validFields[opts.SortBy] {
			sortBy = opts.SortBy
		}
	}

	sortOrder := "DESC"
	if opts.SortOrder != "" && opts.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(` ORDER BY %s %s`, sortBy, sortOrder)

	rows, err := r.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating todos: %w", err)
	}

	// Check for errors from closing rows
	if err = rows.Close(); err != nil {
		return nil, fmt.Errorf("failed to close rows: %w", err)
	}

	return todos, nil
}

// GetByID returns a todo by ID
func (r *TodoRepository) GetByID(id int64) (*models.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = ?
	`

	var todo models.Todo
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return &todo, nil
}

// Update updates a todo
func (r *TodoRepository) Update(id int64, req models.UpdateTodoRequest) (*models.Todo, error) {
	// First, get the existing todo
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	// Build the update query dynamically
	query := "UPDATE todos SET updated_at = ?"
	args := []interface{}{time.Now()}

	if req.Title != nil {
		query += ", title = ?"
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		query += ", description = ?"
		args = append(args, *req.Description)
	}
	if req.Completed != nil {
		query += ", completed = ?"
		args = append(args, *req.Completed)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err = r.db.ExecContext(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	// Return the updated todo
	return r.GetByID(id)
}

// Delete deletes a todo by ID
func (r *TodoRepository) Delete(id int64) error {
	query := "DELETE FROM todos WHERE id = ?"
	result, err := r.db.ExecContext(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
