package main

import (
	"log"

	"projectpi-backend/internal/handlers"
	"projectpi-backend/internal/models"
	"projectpi-backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Initialize DB
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&models.User{}, &models.Song{}, &models.Playlist{}, &models.PlaylistSong{})
	models.DB = db // Set the global DB for models

	// Initialize services and handlers
	userService := &services.UserService{DB: db}
	authHandler := &handlers.AuthHandler{UserService: userService}

	// Setup Gin
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"} // Your frontend URL
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Routes
	r.POST("/signup", authHandler.Signup)
	r.POST("/signin", authHandler.Signin)

	// Protected routes
	protected := r.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		// Song routes
		protected.POST("/upload", handlers.UploadSongHandler)
		protected.GET("/songs", handlers.ListSongsHandler)
		protected.GET("/stream/:id", handlers.StreamSongHandler)
		protected.DELETE("/song/:id", handlers.DeleteSongHandler)
		protected.PUT("/song/:id", handlers.UpdateSongHandler)
		protected.GET("/search", handlers.SearchSongsHandler)

		// Playlist routes
		protected.POST("/playlists", handlers.CreatePlaylistHandler)
		protected.GET("/playlists", handlers.ListPlaylistsHandler)
		protected.GET("/playlist/:id", handlers.GetPlaylistHandler)
		protected.PUT("/playlist/:id", handlers.UpdatePlaylistHandler)
		protected.DELETE("/playlist/:id", handlers.DeletePlaylistHandler)
		protected.POST("/playlist/:id/songs", handlers.AddSongToPlaylistHandler)
		protected.DELETE("/playlist/:id/songs/:songId", handlers.RemoveSongFromPlaylistHandler)
	}

	// Start server
	r.Run(":8080")
}
