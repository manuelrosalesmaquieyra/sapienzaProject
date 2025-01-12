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
	ID          string
	LastMessage string
	Timestamp   time.Time
}
