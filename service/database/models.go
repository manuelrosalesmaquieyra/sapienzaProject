package database

import (
	"database/sql"
	"time"
)

// User representa la estructura de un usuario en la base de datos
type User struct {
	ID       string
	Username string
	Token    string
	PhotoURL string
}

type Conversation struct {
	ID                 string    `json:"conversation_id"`
	LastMessage        string    `json:"last_message"`
	LastMessageIsReply bool      `json:"last_message_is_reply"`
	Timestamp          time.Time `json:"timestamp"`
	Participants       []string  `json:"participants"`
	PhotoURL           string    `json:"photo_url,omitempty"`
	IsGroup            bool      `json:"is_group"`
	Name               string    `json:"name,omitempty"`
}

// Reaction representa una reacción a un mensaje
type Reaction struct {
	UserID   string `json:"user_id" bson:"userID"`
	Reaction string `json:"reaction" bson:"reaction"`
}

type Message struct {
	ID             string         `json:"message_id"`
	ConversationID string         `json:"conversation_id"`
	Sender         string         `json:"sender"`
	Content        sql.NullString `json:"-"`       // Use sql.NullString for nullable fields
	ContentStr     string         `json:"content"` // This will be populated from Content
	ImageURL       sql.NullString `json:"-"`
	ImageURLStr    string         `json:"image_url"`
	ReplyToID      sql.NullString `json:"-"`
	ReplyToIDStr   string         `json:"reply_to_id"`
	Time           time.Time      `json:"timestamp"`
	Reactions      []Reaction     `json:"reactions,omitempty" bson:"reactions,omitempty"`
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
