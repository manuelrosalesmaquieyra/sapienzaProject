package database

import (
	"testing"
	"time"
)

func TestGetMessageByID(t *testing.T) {
	db := setupTestDB(t)

	// Crear un mensaje de prueba
	_, err := db.(*appdbimpl).c.Exec(`
		INSERT INTO messages (id, conversation_id, sender, content, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, "msg1", "conv1", "user1", "Hello, World!", time.Now())
	if err != nil {
		t.Fatalf("error creating test message: %v", err)
	}

	tests := []struct {
		name        string
		messageID   string
		expectError bool
	}{
		{
			name:        "valid message",
			messageID:   "msg1",
			expectError: false,
		},
		{
			name:        "invalid message",
			messageID:   "invalid_id",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := db.GetMessageByID(tt.messageID)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && msg.ID != tt.messageID {
				t.Errorf("expected message ID %s, got %s", tt.messageID, msg.ID)
			}
		})
	}
}

func TestDeleteMessage(t *testing.T) {
	db := setupTestDB(t)

	// Crear un mensaje de prueba
	_, err := db.(*appdbimpl).c.Exec(`
		INSERT INTO messages (id, conversation_id, sender, content, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, "msg1", "conv1", "user1", "Hello, World!", time.Now())
	if err != nil {
		t.Fatalf("error creating test message: %v", err)
	}

	err = db.DeleteMessage("msg1")
	if err != nil {
		t.Errorf("unexpected error deleting message: %v", err)
	}

	// Verificar que el mensaje fue eliminado
	_, err = db.GetMessageByID("msg1")
	if err == nil {
		t.Error("expected error but got none, message was not deleted")
	}
}

func TestForwardMessage(t *testing.T) {
	db := setupTestDB(t)

	// Crear un mensaje de prueba
	_, err := db.(*appdbimpl).c.Exec(`
		INSERT INTO messages (id, conversation_id, sender, content, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, "msg1", "conv1", "user1", "Hello, World!", time.Now())
	if err != nil {
		t.Fatalf("error creating test message: %v", err)
	}

	// Crear una conversación de destino
	_, err = db.(*appdbimpl).c.Exec(`
		INSERT INTO conversations (id, timestamp)
		VALUES (?, ?)
	`, "conv2", time.Now())
	if err != nil {
		t.Fatalf("error creating destination conversation: %v", err)
	}

	// Reenviar el mensaje
	newMsg, err := db.ForwardMessage("msg1", "conv2", "user1")
	if err != nil {
		t.Errorf("unexpected error forwarding message: %v", err)
	}

	if newMsg.ConversationID != "conv2" {
		t.Errorf("expected conversation ID %s, got %s", "conv2", newMsg.ConversationID)
	}

	if newMsg.Content != "Hello, World!" {
		t.Errorf("expected content 'Hello, World!', got %s", newMsg.Content)
	}
}
