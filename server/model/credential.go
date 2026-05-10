package model

import "time"

type Credential struct {
	ID           string
	Name         string
	AccessToken  string
	RefreshToken string
	CreatedAt    time.Time
}
