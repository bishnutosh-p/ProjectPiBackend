package services

import (
	"context"
	"time"

	"projectpi-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlaylistService struct {
	DB *mongo.Database
}

func (s *PlaylistService) CreatePlaylist(playlist *models.Playlist) error {
	collection := s.DB.Collection("playlists")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	playlist.CreatedAt = time.Now()
	playlist.UpdatedAt = time.Now()
	_, err := collection.InsertOne(ctx, playlist)
	return err
}

func (s *PlaylistService) GetPlaylistsByUserID(userID string) ([]models.Playlist, error) {
	collection := s.DB.Collection("playlists")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var playlists []models.Playlist
	err = cursor.All(ctx, &playlists)
	return playlists, err
}

func (s *PlaylistService) GetPlaylistByID(playlistID string) (*models.Playlist, error) {
	collection := s.DB.Collection("playlists")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var playlist models.Playlist
	err := collection.FindOne(ctx, bson.M{"playlist_id": playlistID}).Decode(&playlist)
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

func (s *PlaylistService) UpdatePlaylist(playlistID string, updates bson.M) error {
	collection := s.DB.Collection("playlists")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updates["updated_at"] = time.Now()
	_, err := collection.UpdateOne(ctx, bson.M{"playlist_id": playlistID}, bson.M{"$set": updates})
	return err
}

func (s *PlaylistService) DeletePlaylist(playlistID string) error {
	collection := s.DB.Collection("playlists")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"playlist_id": playlistID})
	if err != nil {
		return err
	}

	// Also delete all songs in this playlist
	playlistSongsCollection := s.DB.Collection("playlist_songs")
	_, err = playlistSongsCollection.DeleteMany(ctx, bson.M{"playlist_id": playlistID})
	return err
}

func (s *PlaylistService) AddSongToPlaylist(playlistID string, songID string, position int) error {
	collection := s.DB.Collection("playlist_songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	playlistSong := models.PlaylistSong{
		PlaylistID: playlistID,
		SongID:     songID,
		Position:   position,
		CreatedAt:  time.Now(),
	}

	_, err := collection.InsertOne(ctx, playlistSong)
	return err
}

func (s *PlaylistService) RemoveSongFromPlaylist(playlistID string, songID string) error {
	collection := s.DB.Collection("playlist_songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"playlist_id": playlistID, "song_id": songID})
	return err
}

func (s *PlaylistService) GetPlaylistSongs(playlistID string) ([]models.PlaylistSong, error) {
	collection := s.DB.Collection("playlist_songs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"playlist_id": playlistID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var playlistSongs []models.PlaylistSong
	err = cursor.All(ctx, &playlistSongs)
	return playlistSongs, err
}
