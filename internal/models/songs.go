package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Song struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	SongID    string             `bson:"song_id"`
	Title     string             `bson:"title"`
	Artist    string             `bson:"artist"`
	Filename  string             `bson:"filename"`
	UserID    string             `bson:"user_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}
