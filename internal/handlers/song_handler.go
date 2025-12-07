package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"projectpi-backend/internal/models"
	"projectpi-backend/internal/services"
	"projectpi-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func UploadSongHandler(c *gin.Context, songService *services.SongService) {
	title := c.PostForm("title")
	artist := c.PostForm("artist")

	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return
	}

	// Validate file size (e.g., max 50MB)
	const maxFileSize = 50 << 20 // 50MB
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 50MB)"})
		return
	}

	// Validate file type (allow only mp3, wav, jpg, png)
	allowedTypes := map[string]bool{
		"audio/mpeg": true,
		"audio/wav":  true,
		"image/jpeg": true,
		"image/png":  true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}

	// Create user-specific directory: uploads/<user_id>/
	userDir := filepath.Join("uploads", userIDStr)
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user directory"})
		return
	}

	// Save file in user-specific directory
	filePath := filepath.Join(userDir, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	songID := utils.GenerateSongID(uint(time.Now().UnixNano() % 10000))
	song := models.Song{
		SongID:   songID,
		Title:    title,
		Artist:   artist,
		Filename: file.Filename,
		UserID:   userIDStr,
	}

	if err := songService.CreateSong(&song); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save song"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song uploaded successfully", "song_id": song.SongID})
}

func ListSongsHandler(c *gin.Context, songService *services.SongService) {
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	songs, err := songService.GetSongsByUserID(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"songs": songs})
}

func StreamSongHandler(c *gin.Context, songService *services.SongService) {
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	songID := c.Param("id")
	song, err := songService.GetSongByID(songID)
	if err != nil || song.UserID != userIDStr {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	userDir := filepath.Join("uploads", userIDStr)
	filePath := filepath.Join(userDir, song.Filename)
	c.File(filePath)
}

func DeleteSongHandler(c *gin.Context, songService *services.SongService) {
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	songID := c.Param("id")
	song, err := songService.GetSongByID(songID)
	if err != nil || song.UserID != userIDStr {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	// Delete file from disk
	userDir := filepath.Join("uploads", userIDStr)
	filePath := filepath.Join(userDir, song.Filename)
	os.Remove(filePath)

	// Delete from DB
	if err := songService.DeleteSong(songID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete song"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song deleted"})
}

func UpdateSongHandler(c *gin.Context, songService *services.SongService) {
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	songID := c.Param("id")
	song, err := songService.GetSongByID(songID)
	if err != nil || song.UserID != userIDStr {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	var input struct {
		Title  string `json:"title"`
		Artist string `json:"artist"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := bson.M{
		"title":  input.Title,
		"artist": input.Artist,
	}

	if err := songService.UpdateSong(songID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update song"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song updated"})
}

func SearchSongsHandler(c *gin.Context, songService *services.SongService) {
	userID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	songs, err := songService.SearchSongs(query, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search songs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"songs": songs})
}
