package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func runMigrations(db *pgxpool.Pool) error {
	// Find migration files
	migrationsDir := findMigrationsDir()
	if migrationsDir == "" {
		log.Println("⚠️  No migrations directory found, skipping migrations")
		return nil
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	// Filter and sort .up.sql files
	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	if len(upFiles) == 0 {
		log.Println("⚠️  No migration files found")
		return nil
	}

	// Create migrations tracking table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	// Run each migration
	for _, file := range upFiles {
		version := strings.TrimSuffix(file, ".up.sql")

		// Check if already applied
		var count int
		err := db.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("checking migration %s: %w", version, err)
		}
		if count > 0 {
			continue
		}

		// Read and execute migration
		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", file, err)
		}

		_, err = db.Exec(context.Background(), string(content))
		if err != nil {
			return fmt.Errorf("executing migration %s: %w", file, err)
		}

		// Record migration
		_, err = db.Exec(context.Background(),
			"INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			return fmt.Errorf("recording migration %s: %w", version, err)
		}

		fmt.Printf("✅ Applied migration: %s\n", file)
	}

	return nil
}

func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
		"backend/migrations",
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}
