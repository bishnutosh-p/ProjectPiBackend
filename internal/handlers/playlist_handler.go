package handlers

import (
	"fmt"
	"net/http"
	"time"

	"projectpi-backend/internal/models"
	"projectpi-backend/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// CreatePlaylistHandler creates a new playlist
func CreatePlaylistHandler(c *gin.Context, playlistService *services.PlaylistService) {
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

	if err := playlistService.CreatePlaylist(&playlist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
		return
	}

	c.JSON(http.StatusCreated, playlist)
}

// ListPlaylistsHandler lists all playlists for the authenticated user
func ListPlaylistsHandler(c *gin.Context, playlistService *services.PlaylistService) {
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	playlists, err := playlistService.GetPlaylistsByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playlists"})
		return
	}

	c.JSON(http.StatusOK, playlists)
}

// GetPlaylistHandler retrieves a specific playlist with its songs
func GetPlaylistHandler(c *gin.Context, playlistService *services.PlaylistService) {
	playlistID := c.Param("id")

	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	playlist, err := playlistService.GetPlaylistByID(playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check if playlist belongs to user
	if playlist.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get songs in playlist
	playlistSongs, err := playlistService.GetPlaylistSongs(playlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playlist songs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist": playlist,
		"songs":    playlistSongs,
	})
}

// UpdatePlaylistHandler updates a playlist's name or description
func UpdatePlaylistHandler(c *gin.Context, playlistService *services.PlaylistService) {
	playlistID := c.Param("id")

	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	playlist, err := playlistService.GetPlaylistByID(playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check ownership
	if playlist.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
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

	updates := bson.M{}
	if request.Name != "" {
		updates["name"] = request.Name
	}
	if request.Description != "" {
		updates["description"] = request.Description
	}

	if err := playlistService.UpdatePlaylist(playlistID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist updated"})
}

// DeletePlaylistHandler deletes a playlist and all its songs
func DeletePlaylistHandler(c *gin.Context, playlistService *services.PlaylistService) {
	playlistID := c.Param("id")

	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	playlist, err := playlistService.GetPlaylistByID(playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check ownership
	if playlist.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := playlistService.DeletePlaylist(playlistID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist deleted"})
}

// AddSongToPlaylistHandler adds a song to a playlist
func AddSongToPlaylistHandler(c *gin.Context, playlistService *services.PlaylistService) {
	playlistID := c.Param("id")

	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	playlist, err := playlistService.GetPlaylistByID(playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check ownership
	if playlist.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var request struct {
		SongID   string `json:"song_id" binding:"required"`
		Position int    `json:"position"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current songs to determine position if not specified
	if request.Position == 0 {
		playlistSongs, _ := playlistService.GetPlaylistSongs(playlistID)
		request.Position = len(playlistSongs) + 1
	}

	if err := playlistService.AddSongToPlaylist(playlistID, request.SongID, request.Position); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add song to playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song added to playlist"})
}

// RemoveSongFromPlaylistHandler removes a song from a playlist
func RemoveSongFromPlaylistHandler(c *gin.Context, playlistService *services.PlaylistService) {
	playlistID := c.Param("id")
	songID := c.Param("songId")

	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	playlist, err := playlistService.GetPlaylistByID(playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check ownership
	if playlist.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := playlistService.RemoveSongFromPlaylist(playlistID, songID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove song from playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song removed from playlist"})
}
