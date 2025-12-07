package services

import (
	"context"
	"time"

	"projectpi-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SongService struct {
	DB *mongo.Database
}

func (s *SongService) CreateSong(song *models.Song) error {
	collection := s.DB.Collection("songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	song.CreatedAt = time.Now()
	song.UpdatedAt = time.Now()
	_, err := collection.InsertOne(ctx, song)
	return err
}

func (s *SongService) GetSongsByUserID(userID string) ([]models.Song, error) {
	collection := s.DB.Collection("songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []models.Song
	err = cursor.All(ctx, &songs)
	return songs, err
}

func (s *SongService) GetSongByID(songID string) (*models.Song, error) {
	collection := s.DB.Collection("songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var song models.Song
	err := collection.FindOne(ctx, bson.M{"song_id": songID}).Decode(&song)
	if err != nil {
		return nil, err
	}
	return &song, nil
}

func (s *SongService) DeleteSong(songID string) error {
	collection := s.DB.Collection("songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"song_id": songID})
	return err
}

func (s *SongService) UpdateSong(songID string, updates bson.M) error {
	collection := s.DB.Collection("songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updates["updated_at"] = time.Now()
	_, err := collection.UpdateOne(ctx, bson.M{"song_id": songID}, bson.M{"$set": updates})
	return err
}

func (s *SongService) SearchSongs(query string, userID string) ([]models.Song, error) {
	collection := s.DB.Collection("songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"artist": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []models.Song
	err = cursor.All(ctx, &songs)
	return songs, err
}
