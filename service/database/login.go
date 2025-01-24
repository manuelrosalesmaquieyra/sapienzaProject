package database

import (
	"errors"
	"fmt"
	"regexp"

	"log"

	"github.com/google/uuid"
)

func (db *appdbimpl) CreateSession(name string) (*Session, error) {
	log.Printf("Creating session for: %s", name)

	// Validate name format
	if len(name) < 3 || len(name) > 16 {
		return nil, errors.New("name must be between 3 and 16 characters")
	}

	namePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !namePattern.MatchString(name) {
		return nil, errors.New("invalid name format")
	}

	// Generate new token
	newToken := uuid.New().String()

	// Check if user exists
	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", name).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error checking user existence: %w", err)
	}

	if exists {
		// Update existing user's token
		_, err = db.c.Exec("UPDATE users SET token = ? WHERE username = ?", newToken, name)
		if err != nil {
			return nil, fmt.Errorf("error updating user token: %w", err)
		}
	} else {
		// Create new user with username as ID
		_, err = db.c.Exec("INSERT INTO users (id, username, token) VALUES (?, ?, ?)",
			name, name, newToken) // Using name as ID
		if err != nil {
			return nil, fmt.Errorf("error creating user: %w", err)
		}
	}

	return &Session{
		Username:   name,
		Identifier: newToken,
	}, nil
}
