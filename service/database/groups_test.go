package database

import (
	"testing"
)

func TestCreateGroup(t *testing.T) {
	db := setupTestDB(t)

	// Crear usuario de prueba
	_, err := db.(*appdbimpl).c.Exec(`
        INSERT INTO users (id, username, token) VALUES (?, ?, ?)
    `, "user1", "testuser", "testtoken123")
	if err != nil {
		t.Fatalf("error creating test user: %v", err)
	}

	tests := []struct {
		name        string
		groupName   string
		creatorID   string
		expectError bool
	}{
		{
			name:        "valid group",
			groupName:   "Test Group",
			creatorID:   "user1",
			expectError: false,
		},
		{
			name:        "empty name",
			groupName:   "",
			creatorID:   "user1",
			expectError: true,
		},
		{
			name:        "invalid creator",
			groupName:   "Test Group",
			creatorID:   "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := db.CreateGroup(tt.groupName, tt.creatorID)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && group == nil {
				t.Error("expected group but got nil")
			}
			if !tt.expectError && group.Name != tt.groupName {
				t.Errorf("expected group name %s, got %s", tt.groupName, group.Name)
			}
		})
	}
}

func TestUpdateGroupName(t *testing.T) {
	db := setupTestDB(t)

	// Crear grupo de prueba
	group, err := db.CreateGroup("Test Group", "user1")
	if err != nil {
		t.Fatalf("error creating test group: %v", err)
	}

	tests := []struct {
		name        string
		groupID     string
		newName     string
		expectError bool
	}{
		{
			name:        "valid update",
			groupID:     group.ID,
			newName:     "New Name",
			expectError: false,
		},
		{
			name:        "empty name",
			groupID:     group.ID,
			newName:     "",
			expectError: true,
		},
		{
			name:        "invalid group",
			groupID:     "invalid",
			newName:     "New Name",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateGroupName(tt.groupID, tt.newName)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUpdateGroupPhoto(t *testing.T) {
	db := setupTestDB(t)

	// Crear grupo de prueba
	group, err := db.CreateGroup("Test Group", "user1")
	if err != nil {
		t.Fatalf("error creating test group: %v", err)
	}

	tests := []struct {
		name        string
		groupID     string
		photoURL    string
		expectError bool
	}{
		{
			name:        "valid update",
			groupID:     group.ID,
			photoURL:    "http://example.com/photo.jpg",
			expectError: false,
		},
		{
			name:        "empty URL",
			groupID:     group.ID,
			photoURL:    "",
			expectError: true,
		},
		{
			name:        "invalid group",
			groupID:     "invalid",
			photoURL:    "http://example.com/photo.jpg",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateGroupPhoto(tt.groupID, tt.photoURL)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLeaveGroup(t *testing.T) {
	db := setupTestDB(t)

	// Crear grupo de prueba
	group, err := db.CreateGroup("Test Group", "user1")
	if err != nil {
		t.Fatalf("error creating test group: %v", err)
	}

	tests := []struct {
		name        string
		groupID     string
		userID      string
		expectError bool
	}{
		{
			name:        "valid leave",
			groupID:     group.ID,
			userID:      "user1",
			expectError: false,
		},
		{
			name:        "invalid group",
			groupID:     "invalid",
			userID:      "user1",
			expectError: true,
		},
		{
			name:        "non-member user",
			groupID:     group.ID,
			userID:      "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.LeaveGroup(tt.groupID, tt.userID)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
