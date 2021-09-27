package models

import (
	"time"
)

type User struct {
	ID         uint
	Name       string
	Company    string
	Email      string
	Permission string
	Status     string
	UpdatedAt  time.Time
	CreatedAt  time.Time
}
