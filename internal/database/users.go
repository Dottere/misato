package database

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    string
}

var ErrInvalidCredentials = errors.New("invalid username or password")

// RegisterUser biztonságosan hasheli a jelszót és elmenti az új felhasználót.
func (db *DB) RegisterUser(username, password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `INSERT INTO users (username, password_hash) VALUES (?, ?)`
	_, err = db.Conn.Exec(query, username, string(hashedBytes))

	if err != nil {
		return fmt.Errorf("failed to register user (username might be taken): %w", err)
	}

	return nil
}

// AuthenticateUser ellenőrzi a belépési adatokat.
// Sikeres belépés esetén visszaadja a felhasználó ID-jét.
func (db *DB) AuthenticateUser(username, password string) (int, error) {
	query := `SELECT id, password_hash FROM users WHERE username = ?`

	var id int
	var storedHash string

	err := db.Conn.QueryRow(query, username).Scan(&id, &storedHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials // Nem létezik ilyen felhasználó
		}
		return 0, fmt.Errorf("database query error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		return 0, ErrInvalidCredentials // A jelszó hibás
	}

	return id, nil
}

func (db *DB) DeleteUser(username, password string) error {
	searchQuery := `SELECT id, password_hash FROM users WHERE username = ?`

	var id int
	var storedHash string

	err := db.Conn.QueryRow(searchQuery, username).Scan(&id, &storedHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials // Nem létezik ilyen felhasználó
		}
		return fmt.Errorf("database query error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		return ErrInvalidCredentials // A jelszó hibás
	}

	deleteQuery := `DELETE FROM users WHERE id = ?`

	_, deletionError := db.Conn.Exec(deleteQuery, id)
	if deletionError != nil {
		return fmt.Errorf("delete failed: %w", deletionError)
	}

	return nil
}
