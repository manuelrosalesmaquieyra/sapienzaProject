package database

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// CreateGroup crea un nuevo grupo
func (db *appdbimpl) CreateGroup(name string, creatorID string) (*Group, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}

	// Verificar que el usuario existe
	var exists bool
	err := db.c.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM users 
            WHERE id = ?
        )
    `, creatorID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error checking user existence: %w", err)
	}
	if !exists {
		return nil, errors.New("user not found")
	}

	group := &Group{
		ID:        generateUUID(),
		Name:      name,
		CreatedAt: time.Now(),
	}

	tx, err := db.c.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("error rolling back transaction: %v", err)
		}
	}()

	// Insertar grupo
	_, err = tx.Exec(`        INSERT INTO groups (id, name, created_at)
        VALUES (?, ?, ?)
    `, group.ID, group.Name, group.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error creating group: %w", err)
	}

	// Añadir creador como miembro
	_, err = tx.Exec(`        INSERT INTO group_members (group_id, user_id)
        VALUES (?, ?)
    `, group.ID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("error adding group creator: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return group, nil
}

// UpdateGroupName actualiza el nombre del grupo
func (db *appdbimpl) UpdateGroupName(groupID string, newName string) error {
	if newName == "" {
		return errors.New("group name is required")
	}

	result, err := db.c.Exec(`
        UPDATE groups 
        SET name = ?
        WHERE id = ?
    `, newName, groupID)
	if err != nil {
		return fmt.Errorf("error updating group name: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %w", err)
	}
	if rows == 0 {
		return errors.New("group not found")
	}

	return nil
}

// UpdateGroupPhoto actualiza la foto del grupo
func (db *appdbimpl) UpdateGroupPhoto(groupID string, photoURL string) error {
	if photoURL == "" {
		return errors.New("photo URL is required")
	}

	result, err := db.c.Exec(`
        UPDATE groups 
        SET photo_url = ?
        WHERE id = ?
    `, photoURL, groupID)
	if err != nil {
		return fmt.Errorf("error updating group photo: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %w", err)
	}
	if rows == 0 {
		return errors.New("group not found")
	}

	return nil
}

// LeaveGroup permite a un usuario abandonar un grupo
func (db *appdbimpl) LeaveGroup(groupID string, userID string) error {
	result, err := db.c.Exec(`
        DELETE FROM group_members
        WHERE group_id = ? AND user_id = ?
    `, groupID, userID)
	if err != nil {
		return fmt.Errorf("error leaving group: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %w", err)
	}
	if rows == 0 {
		return errors.New("user is not a member of this group")
	}

	return nil
}
