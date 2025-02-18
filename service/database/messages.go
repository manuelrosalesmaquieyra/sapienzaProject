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
	// Get the original message with both content and image_url
	var originalMsg Message
	err := db.c.QueryRow(`
        SELECT content, image_url
        FROM messages
        WHERE id = ?
    `, messageID).Scan(&originalMsg.Content, &originalMsg.ImageURL)

	if err != nil {
		return nil, fmt.Errorf("error getting original message: %w", err)
	}

	// Create new message
	newMsg := Message{
		ID:             generateUUID(),
		ConversationID: newConversationID,
		Sender:         senderID,
		Content:        originalMsg.Content,
		ImageURL:       originalMsg.ImageURL,
		Time:           time.Now(),
	}

	// Insert the forwarded message
	_, err = db.c.Exec(`
        INSERT INTO messages (id, conversation_id, sender, content, image_url, timestamp)
        VALUES (?, ?, ?, ?, ?, ?)
    `, newMsg.ID, newMsg.ConversationID, newMsg.Sender, newMsg.Content, newMsg.ImageURL, newMsg.Time)

	if err != nil {
		return nil, fmt.Errorf("error forwarding message: %w", err)
	}

	// Update conversation's last message
	var lastMessage string
	if newMsg.Content.Valid {
		lastMessage = newMsg.Content.String
	} else if newMsg.ImageURL.Valid {
		lastMessage = "[Image]"
	}

	_, err = db.c.Exec(`
        UPDATE conversations 
        SET last_message = ?, timestamp = ?
        WHERE id = ?
    `, lastMessage, newMsg.Time, newConversationID)

	if err != nil {
		return nil, fmt.Errorf("error updating conversation: %w", err)
	}

	// Set the string fields for JSON
	if newMsg.Content.Valid {
		newMsg.ContentStr = newMsg.Content.String
	}
	if newMsg.ImageURL.Valid {
		newMsg.ImageURLStr = newMsg.ImageURL.String
	}

	return &newMsg, nil
}

func (db *appdbimpl) CreateReplyMessage(conversationID, sender, content, replyToID string) (string, error) {
	messageID := generateUUID()

	_, err := db.c.Exec(`
        INSERT INTO messages (id, conversation_id, sender, content, reply_to_id, timestamp)
        VALUES (?, ?, ?, ?, ?, datetime('now'))
    `, messageID, conversationID, sender, content, replyToID)

	if err != nil {
		return "", fmt.Errorf("error creating reply message: %w", err)
	}

	return messageID, nil
}

func (db *appdbimpl) CreateImageMessage(conversationID, sender, imageURL string) (string, error) {
	messageID := generateUUID()

	_, err := db.c.Exec(`
        INSERT INTO messages (id, conversation_id, sender, image_url, timestamp)
        VALUES (?, ?, ?, ?, datetime('now'))
    `, messageID, conversationID, sender, imageURL)

	if err != nil {
		return "", fmt.Errorf("error creating image message: %w", err)
	}

	return messageID, nil
}
