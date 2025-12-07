package main

import (
	"context"
	"log"
	"os"

	"projectpi-backend/internal/handlers"
	"projectpi-backend/internal/services"
	"projectpi-backend/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables from system")
	}

	// Get MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable is not set")
	}

	// Initialize MongoDB
	client, err := utils.InitMongoDB(mongoURI)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatal("Failed to disconnect from MongoDB:", err)
		}
	}()

	// Create database and collections references
	db := client.Database("projectpi")

	// Initialize services
	userService := &services.UserService{DB: db}
	playlistService := &services.PlaylistService{DB: db}
	songService := &services.SongService{DB: db}

	// Initialize handlers
	authHandler := &handlers.AuthHandler{UserService: userService}

	// Setup Gin
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://project-pi-frontend.vercel.app"} // Your frontend URL
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
		protected.POST("/upload", func(c *gin.Context) {
			handlers.UploadSongHandler(c, songService)
		})
		protected.GET("/songs", func(c *gin.Context) {
			handlers.ListSongsHandler(c, songService)
		})
		protected.GET("/stream/:id", func(c *gin.Context) {
			handlers.StreamSongHandler(c, songService)
		})
		protected.DELETE("/song/:id", func(c *gin.Context) {
			handlers.DeleteSongHandler(c, songService)
		})
		protected.PUT("/song/:id", func(c *gin.Context) {
			handlers.UpdateSongHandler(c, songService)
		})
		protected.GET("/search", func(c *gin.Context) {
			handlers.SearchSongsHandler(c, songService)
		})

		// Playlist routes
		protected.POST("/playlists", func(c *gin.Context) {
			handlers.CreatePlaylistHandler(c, playlistService)
		})
		protected.GET("/playlists", func(c *gin.Context) {
			handlers.ListPlaylistsHandler(c, playlistService)
		})
		protected.GET("/playlist/:id", func(c *gin.Context) {
			handlers.GetPlaylistHandler(c, playlistService)
		})
		protected.PUT("/playlist/:id", func(c *gin.Context) {
			handlers.UpdatePlaylistHandler(c, playlistService)
		})
		protected.DELETE("/playlist/:id", func(c *gin.Context) {
			handlers.DeletePlaylistHandler(c, playlistService)
		})
		protected.POST("/playlist/:id/songs", func(c *gin.Context) {
			handlers.AddSongToPlaylistHandler(c, playlistService)
		})
		protected.DELETE("/playlist/:id/songs/:songId", func(c *gin.Context) {
			handlers.RemoveSongFromPlaylistHandler(c, playlistService)
		})
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
