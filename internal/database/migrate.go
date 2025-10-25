package database

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"
)

// Migrator handles database migrations
type Migrator struct {
	db *DB
	fs embed.FS
}

// NewMigrator creates a new Migrator
func NewMigrator(db *DB, fs embed.FS) *Migrator {
	return &Migrator{
		db: db,
		fs: fs,
	}
}

// Run executes all pending migrations
func (m *Migrator) Run() error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	entries, err := m.fs.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files by name
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Get already applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, filename := range migrationFiles {
		if applied[filename] {
			continue
		}

		if err := m.applyMigration(filename); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filename, err)
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL UNIQUE,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := m.db.ExecContext(context.Background(), query)
	return err
}

// getAppliedMigrations returns a set of already applied migration filenames
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	query := "SELECT filename FROM schema_migrations"
	rows, err := m.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	applied := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration file
func (m *Migrator) applyMigration(filename string) error {
	// Read migration file
	content, err := m.fs.ReadFile("migrations/" + filename)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	ctx := context.Background()

	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
			}
		}
	}()

	// Execute migration SQL
	if _, err = tx.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	query := "INSERT INTO schema_migrations (filename) VALUES (?)"
	if _, err = tx.ExecContext(ctx, query, filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Applied migration: %s\n", filename)
	return nil
}
