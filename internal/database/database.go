package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

type MasterToken struct {
	ID          string    `json:"id"`
	Secret      string    `json:"secret"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
	Issuer      *string   `json:"issuer,omitempty"`
	AccountName *string   `json:"account_name,omitempty"`
}

func NewDB() (*DB, error) {
	// Get database configuration from environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "otp_basic")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open database connection
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	db := &DB{conn: conn}

	// Run migrations
	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) runMigrations() error {
	driver, err := postgres.WithInstance(db.conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// MasterToken CRUD operations

func (db *DB) CreateMasterToken(token *MasterToken) error {
	query := `
		INSERT INTO master_tokens (id, secret, created_at, is_active, issuer, account_name)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.conn.Exec(query, token.ID, token.Secret, token.CreatedAt, token.IsActive, token.Issuer, token.AccountName)
	if err != nil {
		return fmt.Errorf("failed to create master token: %w", err)
	}

	return nil
}

func (db *DB) GetMasterToken(id string) (*MasterToken, error) {
	query := `
		SELECT id, secret, created_at, is_active, issuer, account_name
		FROM master_tokens
		WHERE id = $1`

	row := db.conn.QueryRow(query, id)

	token := &MasterToken{}
	err := row.Scan(&token.ID, &token.Secret, &token.CreatedAt, &token.IsActive, &token.Issuer, &token.AccountName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Token not found
		}
		return nil, fmt.Errorf("failed to get master token: %w", err)
	}

	return token, nil
}

func (db *DB) UpdateMasterToken(token *MasterToken) error {
	query := `
		UPDATE master_tokens
		SET secret = $2, is_active = $3, issuer = $4, account_name = $5
		WHERE id = $1`

	_, err := db.conn.Exec(query, token.ID, token.Secret, token.IsActive, token.Issuer, token.AccountName)
	if err != nil {
		return fmt.Errorf("failed to update master token: %w", err)
	}

	return nil
}

func (db *DB) DeleteMasterToken(id string) error {
	query := `DELETE FROM master_tokens WHERE id = $1`

	_, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete master token: %w", err)
	}

	return nil
}

func (db *DB) ListMasterTokens(limit, offset int) ([]*MasterToken, error) {
	query := `
		SELECT id, secret, created_at, is_active, issuer, account_name
		FROM master_tokens
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list master tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*MasterToken
	for rows.Next() {
		token := &MasterToken{}
		err := rows.Scan(&token.ID, &token.Secret, &token.CreatedAt, &token.IsActive, &token.Issuer, &token.AccountName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan master token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
