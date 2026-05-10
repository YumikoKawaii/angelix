package model

import "time"

type Member struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`   // bearer token the wrapper uses
	APIKey    string    `json:"api_key"` // Anthropic API key — never returned to clients
	CreatedAt time.Time `json:"created_at"`
}
