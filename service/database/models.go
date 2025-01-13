package database

import "time"

// User represents the structure of a user in the database
type User struct {
	ID       string
	Username string
	Token    string
	PhotoURL string
}

// Conversation represents a conversation between two or more users
type Conversation struct {
	ID           string    `json:"conversation_id"`
	LastMessage  string    `json:"last_message"`
	Timestamp    time.Time `json:"timestamp"`
	Participants []string  `json:"participants"`
}

// Message represents a message in a conversation
type Message struct {
	ID             string    `json:"message_id"`
	ConversationID string    `json:"conversation_id"`
	Sender         string    `json:"sender"`
	Content        string    `json:"content"`
	Time           time.Time `json:"timestamp"`
}

// Reaction represents a reaction to a message
type Reaction struct {
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
	Reaction  string `json:"reaction"`
}
