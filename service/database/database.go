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

	// Conversation operations
	GetConversationMessages(conversationID string) ([]Message, error)
	SendMessage(conversationID string, senderID string, content string) (*Message, error)
	IsUserInConversation(conversationID string, userID string) (bool, error)

	// Message operations
	GetMessageByID(messageID string) (*Message, error)
	DeleteMessage(messageID string) error
	ForwardMessage(messageID, newConversationID, senderID string) (*Message, error)

	// Reaction operations
	AddReaction(messageID string, userID string, reaction string) error
	RemoveReaction(messageID string, userID string) error

	// Group operations
	CreateGroup(name string, creatorID string) (*Group, error)
	UpdateGroupName(groupID string, newName string) error
	UpdateGroupPhoto(groupID string, photoURL string) error
	LeaveGroup(groupID string, userID string) error

	CreateSession(name string) (*Session, error)

	CreateConversation(participants []string) (string, error)

	CreateMessage(conversationId string, sender string, content string) (string, error)
}

type appdbimpl struct {
	c *sql.DB
}

// New retorna una nueva instancia de AppDatabase
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Crear tablas si no existen
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		token TEXT UNIQUE NOT NULL,
		photo_url TEXT
	);

	CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		last_message TEXT,
		timestamp DATETIME
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		conversation_id TEXT,
		sender TEXT,
		content TEXT,
		timestamp DATETIME,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id)
	);

	CREATE TABLE IF NOT EXISTS conversation_participants (
		conversation_id TEXT,
		user_id TEXT,
		PRIMARY KEY (conversation_id, user_id),
		FOREIGN KEY (conversation_id) REFERENCES conversations(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS reactions (
		message_id TEXT,
		user_id TEXT,
		reaction TEXT,
		PRIMARY KEY (message_id, user_id),
		FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id),
		CHECK (length(reaction) >= 1 AND length(reaction) <= 5)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		identifier TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		FOREIGN KEY (username) REFERENCES users(username)
	);

	CREATE TABLE IF NOT EXISTS groups (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		photo_url TEXT
	);

	CREATE TABLE IF NOT EXISTS group_members (
		group_id TEXT,
		user_id TEXT,
		FOREIGN KEY (group_id) REFERENCES groups(id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		PRIMARY KEY (group_id, user_id)
	);`

	if _, err := db.Exec(sqlStmt); err != nil {
		return nil, fmt.Errorf("error creating database schema: %w", err)
	}

	// // After creating tables, insert test users
	// sqlStmt = `
	// INSERT OR IGNORE INTO users (id, username, token)
	// VALUES
	// 	('user1', 'manuel1', 'token1'),
	// 	('user2', 'manueltest', 'token2');
	// `
	// if _, err := db.Exec(sqlStmt); err != nil {
	// 	return nil, fmt.Errorf("error inserting test users: %w", err)
	// }

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
