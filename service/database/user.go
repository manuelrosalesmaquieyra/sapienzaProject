package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

// GetUserByToken busca un usuario por su token de autenticación
func (db *appdbimpl) GetUserByToken(token string) (*User, error) {
	log.Printf("Searching for user with token: %s", token)

	var user User
	var photoURL sql.NullString // Use sql.NullString for nullable column

	err := db.c.QueryRow(
		"SELECT id, username, token, photo_url FROM users WHERE token = ?",
		token,
	).Scan(&user.ID, &user.Username, &user.Token, &photoURL)

	if err == sql.ErrNoRows {
		log.Printf("No user found with token: %s", token)
		return nil, errors.New("user not found")
	}
	if err != nil {
		log.Printf("Error querying user: %v", err)
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Convert NullString to string
	if photoURL.Valid {
		user.PhotoURL = photoURL.String
	}

	log.Printf("Found user: %s", user.Username)
	return &user, nil
}

// UpdateUsername actualiza el nombre de usuario y todas sus referencias
func (db *appdbimpl) UpdateUsername(userID string, newUsername string) error {
	// Start a transaction
	tx, err := db.c.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if new username already exists
	var exists bool
	err = tx.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = ? AND id != ?)",
		newUsername, userID,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking username existence: %w", err)
	}

	if exists {
		return errors.New("username already taken")
	}

	// Update username in users table
	_, err = tx.Exec(
		"UPDATE users SET username = ? WHERE id = ?",
		newUsername, userID,
	)
	if err != nil {
		return fmt.Errorf("error updating username: %w", err)
	}

	// Update username in messages table (sender field)
	_, err = tx.Exec(
		"UPDATE messages SET sender = ? WHERE sender = ?",
		newUsername, userID,
	)
	if err != nil {
		return fmt.Errorf("error updating message senders: %w", err)
	}

	// Update username in sessions table
	_, err = tx.Exec(
		"UPDATE sessions SET username = ? WHERE username = ?",
		newUsername, userID,
	)
	if err != nil {
		return fmt.Errorf("error updating sessions: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// UpdateUserPhoto actualiza la foto de perfil del usuario
func (db *appdbimpl) UpdateUserPhoto(userID string, photoURL string) error {
	_, err := db.c.Exec("UPDATE users SET photo_url = ? WHERE id = ?", photoURL, userID)
	return err
}
