package database

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

func (db *appdbimpl) CreateSession(name string) (*Session, error) {
	// Validar formato del nombre
	if len(name) < 3 || len(name) > 16 {
		return nil, errors.New("name must be between 3 and 16 characters")
	}

	namePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !namePattern.MatchString(name) {
		return nil, errors.New("invalid name format")
	}

	// Generar identificador único para la sesión
	identifier := uuid.New().String()

	// Verificar si el usuario existe
	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", name).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error checking user existence: %w", err)
	}

	// Si el usuario no existe, crearlo
	if !exists {
		userID := uuid.New().String()
		_, err = db.c.Exec("INSERT INTO users (id, username) VALUES (?, ?)", userID, name)
		if err != nil {
			return nil, fmt.Errorf("error creating user: %w", err)
		}
	}

	// Crear la sesión
	_, err = db.c.Exec("INSERT INTO sessions (identifier, username) VALUES (?, ?)", identifier, name)
	if err != nil {
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	return &Session{
		Identifier: identifier,
	}, nil
}
