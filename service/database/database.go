/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase es la interfaz de alto nivel para la BD
type AppDatabase interface {
	Ping() error

	// User operations
	GetUserByToken(token string) (*User, error)
	UpdateUsername(userID string, newUsername string) error
	UpdateUserPhoto(userID string, photoURL string) error
	GetUserConversations(userID string) ([]Conversation, error)
}

type appdbimpl struct {
	c *sql.DB
}

// New retorna una nueva instancia de AppDatabase
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Crear tabla de usuarios si no existe
	sqlStmt := `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		token TEXT UNIQUE NOT NULL,
		photo_url TEXT
	);`

	if _, err := db.Exec(sqlStmt); err != nil {
		return nil, fmt.Errorf("error creating users table: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
