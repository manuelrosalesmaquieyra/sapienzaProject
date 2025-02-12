package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

// CreateGroup creates a new group with multiple members
func (db *appdbimpl) CreateGroup(name string, creatorID string, members []string) (*Group, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}

	// Start transaction
	tx, err := db.c.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("error rolling back transaction: %v", err)
		}
	}()

	// Create group
	groupID := generateUUID()
	_, err = tx.Exec(`
        INSERT INTO groups (id, name, timestamp)
        VALUES (?, ?, CURRENT_TIMESTAMP)
    `, groupID, name)
	if err != nil {
		return nil, fmt.Errorf("error creating group: %w", err)
	}

	// Add creator as member
	_, err = tx.Exec(`
        INSERT INTO group_members (group_id, user_id)
        VALUES (?, ?)
    `, groupID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("error adding group creator: %w", err)
	}

	// Add other members
	for _, memberUsername := range members {
		// Get user ID from username
		var userID string
		err := tx.QueryRow(`
            SELECT id FROM users WHERE username = ?
        `, memberUsername).Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("error finding user %s: %w", memberUsername, err)
		}

		// Add member to group
		_, err = tx.Exec(`
            INSERT INTO group_members (group_id, user_id)
            VALUES (?, ?)
        `, groupID, userID)
		if err != nil {
			return nil, fmt.Errorf("error adding member %s: %w", memberUsername, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &Group{
		ID:        groupID,
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

// UpdateGroupName updates the group name
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

// UpdateGroupPhoto updates the group photo
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

// LeaveGroup allows a user to leave a group
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
