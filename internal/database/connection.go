package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
	conn *sql.DB
}

func New() (*Database, error) {
	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://qlp_user:qlp_password@localhost:5432/qlp_db?sslmode=disable"
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := conn.Ping(); err != nil {
		log.Printf("⚠️  Database connection failed: %v", err)
		log.Printf("📝 Using in-memory fallback mode")
		// Don't return error - fall back to file-based storage
		return &Database{conn: nil}, nil
	}

	log.Printf("✅ Connected to PostgreSQL database")
	return &Database{conn: conn}, nil
}

func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *Database) IsConnected() bool {
	return db.conn != nil
}

func (db *Database) GetConnection() *sql.DB {
	return db.conn
}