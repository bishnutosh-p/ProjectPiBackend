package models

import "gorm.io/gorm"

type Playlist struct {
	gorm.Model
	PlaylistID  string `gorm:"unique;not null"` // e.g., PLAYLIST-123-...
	Name        string `gorm:"not null"`
	Description string
	UserID      string `gorm:"not null"` // Foreign key to User.UserID
}

type PlaylistSong struct {
	gorm.Model
	PlaylistID string `gorm:"not null"` // Foreign key to Playlist.PlaylistID
	SongID     string `gorm:"not null"` // Foreign key to Song.SongID
	Position   int    `gorm:"not null"` // Order of songs in playlist
}
