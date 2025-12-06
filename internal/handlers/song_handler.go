package handlers

import (
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

	var songs []models.Song
	if err := models.DB.Where("user_id = ?", userIDStr).Find(&songs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"songs": songs})
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
