package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	Conn *sql.DB
}

// Inicializálja a kapcsolatot az SQLite fájllal, és létrehozza a táblákat ha nincsenek.
func InitDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbInstance := &DB{Conn: conn}

	if err := dbInstance.createSchema(); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return dbInstance, nil
}

// createSchema lefuttatja a szükséges SQL parancsokat a táblák létrehozásához
func (db *DB) createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reading_progress (
		user_id INTEGER,
		manga_title TEXT,
		last_page_read INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_title),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err := db.Conn.Exec(schema)
	return err
}
