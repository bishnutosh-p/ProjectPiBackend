package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Playlist struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	PlaylistID  string             `bson:"playlist_id"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	UserID      string             `bson:"user_id"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type PlaylistSong struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	PlaylistID string             `bson:"playlist_id"`
	SongID     string             `bson:"song_id"`
	Position   int                `bson:"position"`
	CreatedAt  time.Time          `bson:"created_at"`
}
