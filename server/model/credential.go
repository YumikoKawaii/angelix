package model

import "time"

type Credential struct {
	ID          string
	Name        string
	AccessToken string
	CreatedAt   time.Time
}
