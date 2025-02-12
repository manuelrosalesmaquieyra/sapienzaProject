package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// GetMessageByID obtiene un mensaje específico por su ID
func (db *appdbimpl) GetMessageByID(messageID string) (*Message, error) {
	var msg Message
	err := db.c.QueryRow(`
        SELECT id, conversation_id, sender, content, timestamp
        FROM messages
        WHERE id = ?
    `, messageID).Scan(&msg.ID, &msg.ConversationID, &msg.Sender, &msg.Content, &msg.Time)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting message: %w", err)
	}

	return &msg, nil
}

// DeleteMessage elimina un mensaje por su ID
func (db *appdbimpl) DeleteMessage(messageID string) error {
	// First verify the message exists and get sender
	var messageOwner string
	err := db.c.QueryRow(`
        SELECT sender 
        FROM messages 
        WHERE id = $1
    `, messageID).Scan(&messageOwner)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("error checking message: %w", err)
	}

	// Delete the message
	result, err := db.c.Exec(`
        DELETE FROM messages
        WHERE id = $1
    `, messageID)

	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("message not found or already deleted")
	}

	return nil
}

// ForwardMessage reenvía un mensaje a otra conversación
func (db *appdbimpl) ForwardMessage(messageID, newConversationID, senderID string) (*Message, error) {
	// Obtener el mensaje original
	originalMsg, err := db.GetMessageByID(messageID)
	if err != nil {
		return nil, err
	}

	// Crear un nuevo mensaje en la nueva conversación
	newMsg := Message{
		ID:             generateUUID(),
		ConversationID: newConversationID,
		Sender:         senderID,
		Content:        originalMsg.Content,
		Time:           time.Now(),
	}

	_, err = db.c.Exec(`
        INSERT INTO messages (id, conversation_id, sender, content, timestamp)
        VALUES (?, ?, ?, ?, ?)
    `, newMsg.ID, newMsg.ConversationID, newMsg.Sender, newMsg.Content, newMsg.Time)

	if err != nil {
		return nil, fmt.Errorf("error forwarding message: %w", err)
	}

	return &newMsg, nil
}
