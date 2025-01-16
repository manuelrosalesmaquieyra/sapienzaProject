package database

import (
	"errors"
	"fmt"
	"log"
	"time"
	"github.com/google/uuid"
)

func generateUUID() string {
	return uuid.New().String()
}

// Función auxiliar para validar que una conversación existe
func (db *appdbimpl) conversationExists(conversationID string) (bool, error) {
	var exists bool
	err := db.c.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM conversations 
            WHERE id = ?
        )
    `, conversationID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking conversation existence: %w", err)
	}
	return exists, nil
}

// GetConversationMessages obtiene los mensajes de una conversación
func (db *appdbimpl) GetConversationMessages(conversationID string) ([]Message, error) {
	// Validar que la conversación existe
	exists, err := db.conversationExists(conversationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("conversation not found")
	}

	rows, err := db.c.Query(`
        SELECT id, sender, content, timestamp
        FROM messages
        WHERE conversation_id = ?
        ORDER BY timestamp DESC
    `, conversationID)
	if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.Sender,
			&msg.Content,
			&msg.Time,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// SendMessage añade un nuevo mensaje a una conversación
func (db *appdbimpl) SendMessage(conversationID string, senderID string, content string) (*Message, error) {
	// Validar que la conversación existe
	conversationExists, err := db.conversationExists(conversationID)
	if err != nil {
		return nil, err
	}
	if !conversationExists {
		return nil, errors.New("conversation not found")
	}

	// Verificar que el sender es parte de la conversación
	var isParticipant bool
	err = db.c.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM conversation_participants 
            WHERE conversation_id = ? AND user_id = ?
        )
    `, conversationID, senderID).Scan(&isParticipant)

	if err != nil {
		return nil, fmt.Errorf("error checking participant: %w", err)
	}
	if !isParticipant {
		return nil, errors.New("sender is not part of the conversation")
	}

		tx, err := db.c.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("error rolling back transaction: %v", err)
		}
	}() // Operación en bases de datos que "deshace" una transacción cuando algo sale mal.

	// Crear mensaje
	msg := Message{
		ID:      generateUUID(),
		Sender:  senderID,
		Content: content,
		Time:    time.Now(),
	}

	// Insertar mensaje
	_, err = tx.Exec(`
        INSERT INTO messages (id, conversation_id, sender, content, timestamp)
        VALUES (?, ?, ?, ?, ?)
    `, msg.ID, conversationID, msg.Sender, msg.Content, msg.Time)
	if err != nil {
		return nil, fmt.Errorf("error inserting message: %w", err)
	}

	// Actualizar último mensaje de la conversación
	_, err = tx.Exec(`
        UPDATE conversations 
        SET last_message = ?, timestamp = ?
        WHERE id = ?
    `, content, msg.Time, conversationID)
	if err != nil {
		return nil, fmt.Errorf("error updating conversation: %w", err)
	}

	// Commit transacción
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &msg, nil
}

// IsUserInConversation verifica si un usuario es participante de una conversación
func (db *appdbimpl) IsUserInConversation(conversationID string, userID string) (bool, error) {
	var isParticipant bool
	err := db.c.QueryRow(`
        SELECT EXISTS(
            SELECT 1 
            FROM conversation_participants 
            WHERE conversation_id = ? AND user_id = ?
        )
    `, conversationID, userID).Scan(&isParticipant)

	if err != nil {
		return false, fmt.Errorf("error checking participant: %w", err)
	}
	return isParticipant, nil
}
