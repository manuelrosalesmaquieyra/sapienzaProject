package database

import (
	"testing"
)

func TestGetUserByToken(t *testing.T) {
	db := setupTestDB(t)

	// Insertar usuario de prueba
	_, err := db.(*appdbimpl).c.Exec(
		"INSERT INTO users (id, username, token) VALUES (?, ?, ?)",
		"test_id", "test_user", "test_token",
	)
	if err != nil {
		t.Fatalf("error inserting test user: %v", err)
	}

	tests := []struct {
		name         string
		token        string
		expectError  bool
		expectedUser *User
	}{
		{
			name:        "valid token",
			token:       "test_token",
			expectError: false,
			expectedUser: &User{
				ID:       "test_id",
				Username: "test_user",
				Token:    "test_token",
			},
		},
		{
			name:        "invalid token",
			token:       "invalid_token",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := db.GetUserByToken(tt.token)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectedUser != nil {
				if user.ID != tt.expectedUser.ID {
					t.Errorf("expected ID %v; got %v", tt.expectedUser.ID, user.ID)
				}
				if user.Username != tt.expectedUser.Username {
					t.Errorf("expected Username %v; got %v", tt.expectedUser.Username, user.Username)
				}
				if user.Token != tt.expectedUser.Token {
					t.Errorf("expected Token %v; got %v", tt.expectedUser.Token, user.Token)
				}
			}
		})
	}
}

func TestUpdateUsername(t *testing.T) {
	db := setupTestDB(t)

	// Insertar usuario inicial
	_, err := db.(*appdbimpl).c.Exec(
		"INSERT INTO users (id, username, token) VALUES (?, ?, ?)",
		"test_id", "old_name", "test_token",
	)
	if err != nil {
		t.Fatalf("error inserting test user: %v", err)
	}

	tests := []struct {
		name        string
		userID      string
		newUsername string
		expectError bool
	}{
		{
			name:        "valid update",
			userID:      "test_id",
			newUsername: "new_name",
			expectError: false,
		},
		{
			name:        "non-existent user",
			userID:      "fake_id",
			newUsername: "new_name",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateUsername(tt.userID, tt.newUsername)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError {
				// Verify the update
				var username string
				err := db.(*appdbimpl).c.QueryRow(
					"SELECT username FROM users WHERE id = ?",
					tt.userID,
				).Scan(&username)

				if err != nil {
					t.Errorf("error verifying update: %v", err)
				}

				if username != tt.newUsername {
					t.Errorf("expected username %v; got %v", tt.newUsername, username)
				}
			}
		})
	}
}

func TestUpdateUserPhoto(t *testing.T) {
	db := setupTestDB(t)

	// Insertar usuario de prueba
	_, err := db.(*appdbimpl).c.Exec(
		"INSERT INTO users (id, username, token) VALUES (?, ?, ?)",
		"test_id", "test_user", "test_token",
	)
	if err != nil {
		t.Fatalf("error inserting test user: %v", err)
	}

	tests := []struct {
		name        string
		userID      string
		photoURL    string
		expectError bool
	}{
		{
			name:        "valid photo update",
			userID:      "test_id",
			photoURL:    "https://example.com/photo.jpg",
			expectError: false,
		},
		{
			name:        "non-existent user",
			userID:      "fake_id",
			photoURL:    "https://example.com/photo.jpg",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateUserPhoto(tt.userID, tt.photoURL)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError {
				// Verify the update
				var photoURL string
				err := db.(*appdbimpl).c.QueryRow(
					"SELECT photo_url FROM users WHERE id = ?",
					tt.userID,
				).Scan(&photoURL)

				if err != nil {
					t.Errorf("error verifying update: %v", err)
				}

				if photoURL != tt.photoURL {
					t.Errorf("expected photo_url %v; got %v", tt.photoURL, photoURL)
				}
			}
		})
	}
}

func TestGetUserConversations(t *testing.T) {
	db := setupTestDB(t)

	// Insertar datos de prueba
	_, err := db.(*appdbimpl).c.Exec(`
		INSERT INTO users (id, username, token) VALUES 
		('user1', 'test_user1', 'token1'),
		('user2', 'test_user2', 'token2')
	`)
	if err != nil {
		t.Fatalf("error inserting test users: %v", err)
	}

	// Insertar conversaciones
	_, err = db.(*appdbimpl).c.Exec(`
		INSERT INTO conversations (id, last_message, timestamp) VALUES 
		('conv1', 'Hello', '2024-01-01 10:00:00'),
		('conv2', 'Hi there', '2024-01-01 11:00:00')
	`)
	if err != nil {
		t.Fatalf("error inserting test conversations: %v", err)
	}

	// Insertar participantes
	_, err = db.(*appdbimpl).c.Exec(`
		INSERT INTO conversation_participants (conversation_id, user_id) VALUES 
		('conv1', 'user1'),
		('conv2', 'user1'),
		('conv2', 'user2')
	`)
	if err != nil {
		t.Fatalf("error inserting test participants: %v", err)
	}

	tests := []struct {
		name        string
		userID      string
		expectCount int
		expectError bool
	}{
		{
			name:        "user with conversations",
			userID:      "user1",
			expectCount: 2,
			expectError: false,
		},
		{
			name:        "user with one conversation",
			userID:      "user2",
			expectCount: 1,
			expectError: false,
		},
		{
			name:        "non-existent user",
			userID:      "fake_user",
			expectCount: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conversations, err := db.GetUserConversations(tt.userID)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(conversations) != tt.expectCount {
				t.Errorf("expected %d conversations; got %d", tt.expectCount, len(conversations))
			}
		})
	}
}
