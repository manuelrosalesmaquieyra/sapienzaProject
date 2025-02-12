package database

import "time"

// User representa la estructura de un usuario en la base de datos
type User struct {
	ID       string
	Username string
	Token    string
	PhotoURL string
}

type Conversation struct {
	ID           string    `json:"conversation_id"`
	LastMessage  string    `json:"last_message"`
	Timestamp    time.Time `json:"timestamp"`
	Participants []string  `json:"participants"`
	PhotoURL     string    `json:"photo_url,omitempty"`
	IsGroup      bool      `json:"is_group"`
	Name         string    `json:"name,omitempty"` // Group name if IsGroup is true
}

// Reaction representa una reacción a un mensaje
type Reaction struct {
	UserID   string `json:"user_id" bson:"userID"`
	Reaction string `json:"reaction" bson:"reaction"`
}

type Message struct {
	ID             string     `json:"message_id"`
	ConversationID string     `json:"conversation_id"`
	Sender         string     `json:"sender"`
	Content        string     `json:"content"`
	Time           time.Time  `json:"timestamp"`
	Reactions      []Reaction `json:"reactions,omitempty" bson:"reactions,omitempty"`
}

// Group representa un grupo de chat
type Group struct {
	ID        string    `json:"group_id"`
	Name      string    `json:"name"`
	PhotoURL  string    `json:"photo_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// GroupMember representa un miembro de un grupo
type GroupMember struct {
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

type Session struct {
	Username   string `json:"username"`
	Identifier string `json:"session_id"`
}

type ConversationDetails struct {
	ID           string   `json:"conversation_id"`
	Participants []string `json:"participants"`
	IsGroup      bool     `json:"is_group"`
	Name         string   `json:"name,omitempty"`
	PhotoURL     string   `json:"photo_url,omitempty"`
}
