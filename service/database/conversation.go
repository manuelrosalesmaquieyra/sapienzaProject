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
        SELECT id, conversation_id, sender, content, timestamp
        FROM messages
        WHERE conversation_id = ?
        ORDER BY timestamp ASC
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
			&msg.ConversationID,
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

// IsUserInConversation checks if a user is part of a conversation
func (db *appdbimpl) IsUserInConversation(username string, conversationId string) (bool, error) {
	var count int
	err := db.c.QueryRow(`
        SELECT COUNT(*) 
        FROM conversation_participants 
        WHERE conversation_id = ? AND user_id = ?`,
		conversationId, username).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("error checking conversation participant: %w", err)
	}

	return count > 0, nil
}

// CreateConversation creates a new conversation between users
func (db *appdbimpl) CreateConversation(participants []string) (string, error) {
	log.Printf("Creating conversation with participants: %v", participants)

	tx, err := db.c.Begin()
	if err != nil {
		return "", fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate conversation ID
	conversationID := generateUUID()
	log.Printf("Generated conversation ID: %s", conversationID)

	// Create conversation with current timestamp
	_, err = tx.Exec(`
        INSERT INTO conversations (id, timestamp, last_message)
        VALUES (?, ?, ?)
    `, conversationID, time.Now(), "")
	if err != nil {
		return "", fmt.Errorf("error creating conversation: %w", err)
	}

	// Add participants
	for _, participant := range participants {
		log.Printf("Adding participant: %s", participant)
		_, err = tx.Exec(`
            INSERT INTO conversation_participants (conversation_id, user_id)
            VALUES (?, ?)
        `, conversationID, participant)
		if err != nil {
			return "", fmt.Errorf("error adding participant: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("error committing transaction: %w", err)
	}

	log.Printf("Successfully created conversation %s with participants %v", conversationID, participants)
	return conversationID, nil
}

// GetUserConversations obtiene todas las conversaciones de un usuario
func (db *appdbimpl) GetUserConversations(username string) ([]Conversation, error) { // Changed parameter to username
	log.Printf("Getting conversations for user: %s", username)

	rows, err := db.c.Query(`
        SELECT DISTINCT c.id, COALESCE(c.last_message, ''), c.timestamp
        FROM conversations c
        JOIN conversation_participants cp ON c.id = cp.conversation_id
        WHERE cp.user_id = ?
        ORDER BY c.timestamp DESC`, username) // Using username directly
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

		// Simplified query to get participants
		pRows, err := db.c.Query(`
            SELECT user_id 
            FROM conversation_participants
            WHERE conversation_id = ?`, conv.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting participants: %w", err)
		}
		defer pRows.Close()

		var participants []string
		for pRows.Next() {
			var p string
			if err := pRows.Scan(&p); err != nil {
				return nil, fmt.Errorf("error scanning participant: %w", err)
			}
			participants = append(participants, p)
		}

		conv.Participants = participants
		log.Printf("Conversation %s has participants: %v", conv.ID, participants)
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (db *appdbimpl) CreateMessage(conversationId string, sender string, content string) (string, error) {
	messageId := generateUUID()

	_, err := db.c.Exec(`
        INSERT INTO messages (id, conversation_id, sender, content, timestamp)
        VALUES (?, ?, ?, ?, ?)`,
		messageId, conversationId, sender, content, time.Now())

	if err != nil {
		return "", fmt.Errorf("error creating message: %w", err)
	}

	// Update last_message in conversation
	_, err = db.c.Exec(`
        UPDATE conversations 
        SET last_message = ?
        WHERE id = ?`,
		content, conversationId)

	if err != nil {
		return "", fmt.Errorf("error updating conversation: %w", err)
	}

	return messageId, nil
}
