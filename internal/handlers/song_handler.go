package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"projectpi-backend/internal/models"
	"projectpi-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func UploadSongHandler(c *gin.Context) {
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

	song := models.Song{
		Title:    title,
		Artist:   artist,
		Filename: file.Filename,
		UserID:   userIDStr,
	}
	if err := models.SaveSong(&song); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save song"})
		return
	}
	// Set SongID after DB ID is available
	song.SongID = utils.GenerateSongID(song.ID)
	models.DB.Save(&song)

	c.JSON(http.StatusOK, gin.H{"message": "Song uploaded successfully", "song_id": song.SongID})
}

func ListSongsHandler(c *gin.Context) {
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

	// Get pagination params
	page := 1
	limit := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	offset := (page - 1) * limit

	var songs []models.Song
	if err := models.DB.Where("user_id = ?", userIDStr).Offset(offset).Limit(limit).Find(&songs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"songs": songs, "page": page, "limit": limit})
}

func StreamSongHandler(c *gin.Context) {
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
	var song models.Song
	if err := models.DB.Where("song_id = ? AND user_id = ?", songID, userIDStr).First(&song).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	userDir := filepath.Join("uploads", userIDStr)
	filePath := filepath.Join(userDir, song.Filename)
	c.File(filePath)
}

func DeleteSongHandler(c *gin.Context) {
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
	var song models.Song
	if err := models.DB.Where("song_id = ? AND user_id = ?", songID, userIDStr).First(&song).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	// Delete file from disk
	userDir := filepath.Join("uploads", userIDStr)
	filePath := filepath.Join(userDir, song.Filename)
	os.Remove(filePath)

	// Delete from DB
	if err := models.DB.Delete(&song).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete song"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song deleted"})
}

func UpdateSongHandler(c *gin.Context) {
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
	var song models.Song
	if err := models.DB.Where("song_id = ? AND user_id = ?", songID, userIDStr).First(&song).Error; err != nil {
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

	song.Title = input.Title
	song.Artist = input.Artist
	if err := models.DB.Save(&song).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update song"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song updated", "song": song})
}
