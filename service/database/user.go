package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// GetUserByToken busca un usuario por su token de autenticación
func (db *appdbimpl) GetUserByToken(token string) (*User, error) {
	var user User
	err := db.c.QueryRow(
		"SELECT id, username, token, photo_url FROM users WHERE token = ?",
		token,
	).Scan(&user.ID, &user.Username, &user.Token, &user.PhotoURL)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

// UpdateUsername actualiza el nombre de usuario
func (db *appdbimpl) UpdateUsername(userID string, newUsername string) error {
	// Primero verificamos si el nuevo username ya existe
	var exists bool
	err := db.c.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = ? AND id != ?)",
		newUsername, userID,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking username existence: %w", err)
	}

	if exists {
		return errors.New("username already taken")
	}

	// Si no existe, actualizamos el username
	result, err := db.c.Exec(
		"UPDATE users SET username = ? WHERE id = ?",
		newUsername, userID,
	)
	if err != nil {
		return fmt.Errorf("error updating username: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %w", err)
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateUserPhoto actualiza la foto de perfil del usuario
func (db *appdbimpl) UpdateUserPhoto(userID string, photoURL string) error {
	result, err := db.c.Exec(
		"UPDATE users SET photo_url = ? WHERE id = ?",
		photoURL, userID,
	)
	if err != nil {
		return fmt.Errorf("error updating photo: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %w", err)
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetUserConversations obtiene todas las conversaciones de un usuario
func (db *appdbimpl) GetUserConversations(userID string) ([]Conversation, error) {
	rows, err := db.c.Query(`
		SELECT c.id, c.last_message, c.timestamp 
		FROM conversations c
		JOIN conversation_participants cp ON c.id = cp.conversation_id
		WHERE cp.user_id = ?
		ORDER BY c.timestamp DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("error getting conversations: %w", err)
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.LastMessage, &conv.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}
