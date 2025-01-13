package database

import (
	"testing"
	"time"
)

func TestAddReaction(t *testing.T) {
	db := setupTestDB(t)

	// Crear usuario y mensaje de prueba
	_, err := db.(*appdbimpl).c.Exec(`
        INSERT INTO users (id, username, token) VALUES (?, ?, ?)
    `, "user1", "testuser", "testtoken123")
	if err != nil {
		t.Fatalf("error creating test user: %v", err)
	}

	_, err = db.(*appdbimpl).c.Exec(`
        INSERT INTO messages (id, conversation_id, sender, content, timestamp)
        VALUES (?, ?, ?, ?, ?)
    `, "msg1", "conv1", "user1", "test message", time.Now())
	if err != nil {
		t.Fatalf("error creating test message: %v", err)
	}

	tests := []struct {
		name        string
		messageID   string
		userID      string
		reaction    string
		expectError bool
	}{
		{
			name:        "valid reaction",
			messageID:   "msg1",
			userID:      "user1",
			reaction:    "<3",
			expectError: false,
		},
		{
			name:        "invalid message",
			messageID:   "invalid",
			userID:      "user1",
			reaction:    "<3",
			expectError: true,
		},
		{
			name:        "reaction too long",
			messageID:   "msg1",
			userID:      "user1",
			reaction:    "toolong",
			expectError: true,
		},
		{
			name:        "empty reaction",
			messageID:   "msg1",
			userID:      "user1",
			reaction:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.AddReaction(tt.messageID, tt.userID, tt.reaction)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRemoveReaction(t *testing.T) {
	db := setupTestDB(t)

	// Crear datos de prueba
	_, err := db.(*appdbimpl).c.Exec(`
        INSERT INTO users (id, username, token) VALUES (?, ?, ?)
    `, "user1", "testuser", "testtoken123")
	if err != nil {
		t.Fatalf("error creating test user: %v", err)
	}

	_, err = db.(*appdbimpl).c.Exec(`
        INSERT INTO messages (id, conversation_id, sender, content, timestamp)
        VALUES (?, ?, ?, ?, ?)
    `, "msg1", "conv1", "user1", "test message", time.Now())
	if err != nil {
		t.Fatalf("error creating test message: %v", err)
	}

	// Añadir una reacción para probar su eliminación
	err = db.AddReaction("msg1", "user1", "<3")
	if err != nil {
		t.Fatalf("error adding test reaction: %v", err)
	}

	tests := []struct {
		name        string
		messageID   string
		userID      string
		expectError bool
	}{
		{
			name:        "valid removal",
			messageID:   "msg1",
			userID:      "user1",
			expectError: false,
		},
		{
			name:        "non-existent reaction",
			messageID:   "msg1",
			userID:      "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.RemoveReaction(tt.messageID, tt.userID)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
