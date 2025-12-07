package handlers

import (
	"fmt"
	"net/http"
	"time"

	"projectpi-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// CreatePlaylistHandler creates a new playlist
func CreatePlaylistHandler(c *gin.Context) {
	var request struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Generate playlist ID
	playlistID := fmt.Sprintf("PLAYLIST-%d-%d", time.Now().Unix(), time.Now().UnixNano()%1000000)

	playlist := models.Playlist{
		PlaylistID:  playlistID,
		Name:        request.Name,
		Description: request.Description,
		UserID:      userID.(string),
	}

	if err := models.DB.Create(&playlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
		return
	}

	c.JSON(http.StatusCreated, playlist)
}

// ListPlaylistsHandler lists all playlists for the authenticated user
func ListPlaylistsHandler(c *gin.Context) {
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var playlists []models.Playlist
	if err := models.DB.Where("user_id = ?", userID).Find(&playlists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playlists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

// GetPlaylistHandler gets a specific playlist with its songs
func GetPlaylistHandler(c *gin.Context) {
	playlistID := c.Param("id")
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var playlist models.Playlist
	if err := models.DB.Where("playlist_id = ? AND user_id = ?", playlistID, userID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Get songs in playlist
	var playlistSongs []models.PlaylistSong
	if err := models.DB.Where("playlist_id = ?", playlistID).Order("position").Find(&playlistSongs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playlist songs"})
		return
	}

	// Fetch song details
	var songs []models.Song
	for _, ps := range playlistSongs {
		var song models.Song
		if err := models.DB.Where("song_id = ?", ps.SongID).First(&song).Error; err == nil {
			songs = append(songs, song)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist": playlist,
		"songs":    songs,
	})
}

// UpdatePlaylistHandler updates playlist metadata
func UpdatePlaylistHandler(c *gin.Context) {
	playlistID := c.Param("id")
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var playlist models.Playlist
	if err := models.DB.Where("playlist_id = ? AND user_id = ?", playlistID, userID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	if request.Name != "" {
		playlist.Name = request.Name
	}
	playlist.Description = request.Description

	if err := models.DB.Save(&playlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update playlist"})
		return
	}

	c.JSON(http.StatusOK, playlist)
}

// DeletePlaylistHandler deletes a playlist
func DeletePlaylistHandler(c *gin.Context) {
	playlistID := c.Param("id")
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var playlist models.Playlist
	if err := models.DB.Where("playlist_id = ? AND user_id = ?", playlistID, userID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Delete playlist songs first
	models.DB.Where("playlist_id = ?", playlistID).Delete(&models.PlaylistSong{})

	// Delete playlist
	if err := models.DB.Delete(&playlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist deleted successfully"})
}

// AddSongToPlaylistHandler adds a song to a playlist
func AddSongToPlaylistHandler(c *gin.Context) {
	playlistID := c.Param("id")
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		SongID string `json:"song_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify playlist belongs to user
	var playlist models.Playlist
	if err := models.DB.Where("playlist_id = ? AND user_id = ?", playlistID, userID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Verify song exists and belongs to user
	var song models.Song
	if err := models.DB.Where("song_id = ? AND user_id = ?", request.SongID, userID).First(&song).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	// Check if song already in playlist
	var existing models.PlaylistSong
	if err := models.DB.Where("playlist_id = ? AND song_id = ?", playlistID, request.SongID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Song already in playlist"})
		return
	}

	// Get max position
	var maxPosition int
	models.DB.Model(&models.PlaylistSong{}).Where("playlist_id = ?", playlistID).Select("COALESCE(MAX(position), 0)").Scan(&maxPosition)

	playlistSong := models.PlaylistSong{
		PlaylistID: playlistID,
		SongID:     request.SongID,
		Position:   maxPosition + 1,
	}

	if err := models.DB.Create(&playlistSong).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add song to playlist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Song added to playlist"})
}

// RemoveSongFromPlaylistHandler removes a song from a playlist
func RemoveSongFromPlaylistHandler(c *gin.Context) {
	playlistID := c.Param("id")
	songID := c.Param("songId")
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify playlist belongs to user
	var playlist models.Playlist
	if err := models.DB.Where("playlist_id = ? AND user_id = ?", playlistID, userID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Delete playlist song
	if err := models.DB.Where("playlist_id = ? AND song_id = ?", playlistID, songID).Delete(&models.PlaylistSong{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove song from playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song removed from playlist"})
}
