package models

import "time"

type User struct {
	ID                int    `json:"id"`
	Email             string `json:"email"`
	EncryptedPassword string `json:"encrypted_password"`
}

type UserDTO struct {
	Email             string `json:"email"`
	EncryptedPassword string `json:"encrypted_password"`
}

type Message struct {
	MessageID int       `json:"message_id"`
	UserID    int       `json:"user_id"`
	Body      string    `json:"body"`
	Time      time.Time `json:"time"`
}

type MessageDTO struct {
	UserID int       `json:"user_id"`
	Body   string    `json:"body"`
	Time   time.Time `json:"time"`
}
