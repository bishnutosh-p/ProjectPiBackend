package models

import "gorm.io/gorm"

type Song struct {
	gorm.Model
	SongID   string `gorm:"unique;not null"` // e.g., SONG-123-...
	Title    string `gorm:"not null"`
	Artist   string
	Filename string `gorm:"not null"`
	UserID   string `gorm:"not null"` // Foreign key to User.UserID
}

var DB *gorm.DB

func SaveSong(song *Song) error {
	if DB == nil {
		return gorm.ErrInvalidDB
	}
	return DB.Create(song).Error
}
