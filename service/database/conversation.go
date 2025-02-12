package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
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
	// Check if it's a regular conversation or a group
	var exists bool
	err := db.c.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM conversations WHERE id = ?
            UNION
            SELECT 1 FROM groups WHERE id = ?
        )
    `, conversationID, conversationID).Scan(&exists)

	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("conversation not found")
	}

	// Get messages with their reactions
	rows, err := db.c.Query(`
        SELECT m.id, m.conversation_id, m.sender, m.content, m.timestamp,
               r.user_id, r.reaction
        FROM messages m
        LEFT JOIN reactions r ON m.id = r.message_id
        WHERE m.conversation_id = ?
        ORDER BY m.timestamp ASC
    `, conversationID)
	if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	defer rows.Close()

	messageMap := make(map[string]*Message)

	for rows.Next() {
		var msg Message
		var userID, reaction sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Sender,
			&msg.Content,
			&msg.Time,
			&userID,
			&reaction,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}

		// Get or create message in map
		existingMsg, exists := messageMap[msg.ID]
		if !exists {
			msg.Reactions = make([]Reaction, 0)
			messageMap[msg.ID] = &msg
			existingMsg = &msg
		}

		// Add reaction if present
		if userID.Valid && reaction.Valid {
			existingMsg.Reactions = append(existingMsg.Reactions, Reaction{
				UserID:   userID.String,
				Reaction: reaction.String,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Convert map to slice
	messages := make([]Message, 0, len(messageMap))
	for _, msg := range messageMap {
		messages = append(messages, *msg)
	}

	// Sort by timestamp
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Time.Before(messages[j].Time)
	})

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
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("error rolling back transaction: %v", err)
		}
	}()

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
        SELECT COUNT(*) FROM (
            -- Check regular conversations
            SELECT cp.conversation_id
            FROM conversation_participants cp
            JOIN users u ON cp.user_id = u.id
            WHERE cp.conversation_id = ? AND u.username = ?
            UNION ALL
            -- Check group conversations
            SELECT gm.group_id
            FROM group_members gm
            JOIN users u ON gm.user_id = u.id
            WHERE gm.group_id = ? AND u.username = ?
        )`,
		conversationId, username, conversationId, username).Scan(&count)

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
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("error rolling back transaction: %v", err)
		}
	}()

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

	// Add participants using their user IDs
	for _, username := range participants {
		// Get user ID for the username
		var userID string
		err := tx.QueryRow(`
            SELECT id FROM users WHERE username = ?
        `, username).Scan(&userID)
		if err != nil {
			return "", fmt.Errorf("error getting user ID for %s: %w", username, err)
		}

		log.Printf("Adding participant: %s (ID: %s)", username, userID)
		_, err = tx.Exec(`
            INSERT INTO conversation_participants (conversation_id, user_id)
            VALUES (?, ?)
        `, conversationID, userID)
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
func (db *appdbimpl) GetUserConversations(username string) ([]Conversation, error) {
	log.Printf("Getting conversations for user: %s", username)

	// First get direct conversations
	query := `
        SELECT DISTINCT 
            c.id, 
            COALESCE(c.last_message, ''), 
            c.timestamp,
            FALSE as is_group,
            '' as group_name,
            COALESCE(u2.photo_url, '') as photo_url
        FROM conversations c
        JOIN conversation_participants cp ON c.id = cp.conversation_id
        JOIN users u ON cp.user_id = u.id
        JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
        JOIN users u2 ON cp2.user_id = u2.id
        WHERE u.username = ? AND u2.username != ?
        UNION ALL
        SELECT 
            g.id,
            '' as last_message,
            g.timestamp,
            TRUE as is_group,
            g.name as group_name,
            COALESCE(g.photo_url, '') as photo_url
        FROM groups g
        JOIN group_members gm ON g.id = gm.group_id
        JOIN users u ON gm.user_id = u.id
        WHERE u.username = ?
        ORDER BY timestamp DESC`

	rows, err := db.c.Query(query, username, username, username)
	if err != nil {
		return nil, fmt.Errorf("error getting conversations: %w", err)
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		var isGroup bool
		var groupName string
		var photoURL string
		err := rows.Scan(
			&conv.ID,
			&conv.LastMessage,
			&conv.Timestamp,
			&isGroup,
			&groupName,
			&photoURL,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}

		conv.IsGroup = isGroup
		conv.PhotoURL = photoURL
		if isGroup {
			conv.Name = groupName
			// Get group members
			members, err := db.getGroupMembers(conv.ID)
			if err != nil {
				return nil, fmt.Errorf("error getting group members: %w", err)
			}
			conv.Participants = members
		} else {
			// Get participants for direct conversations
			participants, err := db.GetConversationParticipants(conv.ID)
			if err != nil {
				return nil, fmt.Errorf("error getting participants: %w", err)
			}
			conv.Participants = participants
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// Helper function to get group members
func (db *appdbimpl) getGroupMembers(groupID string) ([]string, error) {
	rows, err := db.c.Query(`
        SELECT u.username
        FROM group_members gm
        JOIN users u ON gm.user_id = u.id
        WHERE gm.group_id = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		members = append(members, username)
	}
	return members, nil
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

func (db *appdbimpl) GetConversationParticipants(conversationId string) ([]string, error) {
	rows, err := db.c.Query(`
        SELECT u.username
        FROM conversation_participants cp
        JOIN users u ON cp.user_id = u.id
        WHERE cp.conversation_id = ?
    `, conversationId)
	if err != nil {
		return nil, fmt.Errorf("error getting participants: %w", err)
	}
	defer rows.Close()

	var participants []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, fmt.Errorf("error scanning participant: %w", err)
		}
		participants = append(participants, username)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating participants: %w", err)
	}

	return participants, nil
}

func (db *appdbimpl) GetConversationDetails(conversationID string) (*ConversationDetails, error) {
	// First check if this is a group
	var isGroup bool
	//var groupName string
	err := db.c.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM groups WHERE id = ?
        )`, conversationID).Scan(&isGroup)
	if err != nil {
		return nil, fmt.Errorf("error checking if group: %w", err)
	}

	details := &ConversationDetails{
		ID:      conversationID,
		IsGroup: isGroup,
	}

	if isGroup {
		// Get group details
		err = db.c.QueryRow(`
            SELECT name, COALESCE(photo_url, '') 
            FROM groups 
            WHERE id = ?`, conversationID).Scan(&details.Name, &details.PhotoURL)
		if err != nil {
			return nil, fmt.Errorf("error getting group details: %w", err)
		}

		// Get group members
		rows, err := db.c.Query(`
            SELECT u.username 
            FROM group_members gm
            JOIN users u ON gm.user_id = u.id
            WHERE gm.group_id = ?`, conversationID)
		if err != nil {
			return nil, fmt.Errorf("error getting group members: %w", err)
		}
		defer rows.Close()

		var members []string
		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return nil, fmt.Errorf("error scanning member: %w", err)
			}
			members = append(members, username)
		}
		details.Participants = members
	} else {
		// Get regular conversation participants
		rows, err := db.c.Query(`
            SELECT u.username 
            FROM conversation_participants cp
            JOIN users u ON cp.user_id = u.id
            WHERE cp.conversation_id = ?`, conversationID)
		if err != nil {
			return nil, fmt.Errorf("error getting participants: %w", err)
		}
		defer rows.Close()

		var participants []string
		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return nil, fmt.Errorf("error scanning participant: %w", err)
			}
			participants = append(participants, username)
		}
		details.Participants = participants
	}

	return details, nil
}
