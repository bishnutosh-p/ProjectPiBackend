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
	db.AutoMigrate(&models.User{}, &models.Song{})
	models.DB = db // Set the global DB for models

	// Initialize services and handlers
	userService := &services.UserService{DB: db}
	authHandler := &handlers.AuthHandler{UserService: userService}

	// Setup Gin
	r := gin.Default()
	r.Use(cors.Default())

	// Routes
	r.POST("/signup", authHandler.Signup)
	r.POST("/signin", authHandler.Signin)

	// Protected routes
	protected := r.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.POST("/upload", handlers.UploadSongHandler)
		protected.GET("/songs", handlers.ListSongsHandler)
		protected.GET("/stream/:id", handlers.StreamSongHandler)
		protected.DELETE("/song/:id", handlers.DeleteSongHandler)
		protected.PUT("/song/:id", handlers.UpdateSongHandler)
	}

	// Start server
	r.Run(":8080")
}
