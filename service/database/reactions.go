package database

import (
	"errors"
	"fmt"
)

// AddReaction añade una reacción a un mensaje
func (db *appdbimpl) AddReaction(messageID string, userID string, reaction string) error {
	// Validar longitud de la reacción
	if len(reaction) < 1 || len(reaction) > 5 {
		return errors.New("reaction must be between 1 and 5 characters")
	}

	// Verificar que el mensaje existe
	exists, err := db.messageExists(messageID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("message not found")
	}

	// Insertar reacción
	_, err = db.c.Exec(`
        INSERT OR REPLACE INTO reactions (message_id, user_id, reaction)
        VALUES (?, ?, ?)
    `, messageID, userID, reaction)

	if err != nil {
		return fmt.Errorf("error adding reaction: %w", err)
	}

	return nil
}

// RemoveReaction elimina una reacción de un mensaje
func (db *appdbimpl) RemoveReaction(messageID string, userID string) error {
	result, err := db.c.Exec(`
        DELETE FROM reactions
        WHERE message_id = ? AND user_id = ?
    `, messageID, userID)

	if err != nil {
		return fmt.Errorf("error removing reaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("reaction not found")
	}

	return nil
}

// messageExists verifica si un mensaje existe
func (db *appdbimpl) messageExists(messageID string) (bool, error) {
	var exists bool
	err := db.c.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM messages 
            WHERE id = ?
        )
    `, messageID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking message existence: %w", err)
	}
	return exists, nil
}
