package database

import (
	"testing"
	"time"
)

func TestGetConversationMessages(t *testing.T) {
	db := setupTestDB(t)

	// Crear una conversación de prueba
	_, err := db.(*appdbimpl).c.Exec(`
		INSERT INTO conversations (id, last_message, timestamp) 
		VALUES (?, ?, ?)
	`, "conv1", "Hello", time.Now())
	if err != nil {
		t.Fatalf("error creating test conversation: %v", err)
	}

	// Insertar algunos mensajes de prueba
	messages := []struct {
		id      string
		sender  string
		content string
	}{
		{"msg1", "user1", "Hello"},
		{"msg2", "user2", "Hi there"},
	}

	for _, msg := range messages {
		_, err := db.(*appdbimpl).c.Exec(`
			INSERT INTO messages (id, conversation_id, sender, content, timestamp)
			VALUES (?, ?, ?, ?, ?)
		`, msg.id, "conv1", msg.sender, msg.content, time.Now())
		if err != nil {
			t.Fatalf("error creating test message: %v", err)
		}
	}

	tests := []struct {
		name           string
		conversationID string
		expectError    bool
		expectedCount  int
	}{
		{
			name:           "valid conversation",
			conversationID: "conv1",
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "invalid conversation",
			conversationID: "invalid_id",
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := db.GetConversationMessages(tt.conversationID)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(messages) != tt.expectedCount {
				t.Errorf("expected %d messages, got %d", tt.expectedCount, len(messages))
			}
		})
	}
}

func TestSendMessage(t *testing.T) {
	db := setupTestDB(t)

	// Crear conversación y participantes de prueba
	_, err := db.(*appdbimpl).c.Exec(`
		INSERT INTO conversations (id, timestamp) 
		VALUES (?, ?)
	`, "conv1", time.Now())
	if err != nil {
		t.Fatalf("error creating test conversation: %v", err)
	}

	_, err = db.(*appdbimpl).c.Exec(`
		INSERT INTO conversation_participants (conversation_id, user_id)
		VALUES (?, ?)
	`, "conv1", "user1")
	if err != nil {
		t.Fatalf("error adding test participant: %v", err)
	}

	tests := []struct {
		name           string
		conversationID string
		senderID       string
		content        string
		expectError    bool
	}{
		{
			name:           "valid message",
			conversationID: "conv1",
			senderID:       "user1",
			content:        "Hello!",
			expectError:    false,
		},
		{
			name:           "invalid conversation",
			conversationID: "invalid_id",
			senderID:       "user1",
			content:        "Hello!",
			expectError:    true,
		},
		{
			name:           "non-participant sender",
			conversationID: "conv1",
			senderID:       "user2",
			content:        "Hello!",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := db.SendMessage(tt.conversationID, tt.senderID, tt.content)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError {
				if msg == nil {
					t.Fatal("expected message but got nil")
					return
				}
				if msg.Content != tt.content {
					t.Errorf("expected content %s, got %s", tt.content, msg.Content)
				}
				if msg.Sender != tt.senderID {
					t.Errorf("expected sender %s, got %s", tt.senderID, msg.Sender)
				}
			}
		})
	}
}

func TestConversationExists(t *testing.T) {
	db := setupTestDB(t)

	// Crear conversación de prueba
	_, err := db.(*appdbimpl).c.Exec(`
		INSERT INTO conversations (id, timestamp) 
		VALUES (?, ?)
	`, "conv1", time.Now())
	if err != nil {
		t.Fatalf("error creating test conversation: %v", err)
	}

	tests := []struct {
		name           string
		conversationID string
		expectExists   bool
	}{
		{
			name:           "existing conversation",
			conversationID: "conv1",
			expectExists:   true,
		},
		{
			name:           "non-existing conversation",
			conversationID: "invalid_id",
			expectExists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := db.(*appdbimpl).conversationExists(tt.conversationID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if exists != tt.expectExists {
				t.Errorf("expected exists=%v, got %v", tt.expectExists, exists)
			}
		})
	}
}
