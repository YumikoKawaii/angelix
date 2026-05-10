package model

import "time"

type Credential struct {
	ID        string
	Name      string
	APIKey    string
	CreatedAt time.Time
}
