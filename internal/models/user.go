package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserID   string `gorm:"unique;not null"` // e.g., USER-123-...
	Username string `gorm:"unique;not null"`
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}
