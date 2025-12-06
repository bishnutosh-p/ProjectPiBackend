package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
	print("Starting API server...\n")
    r := gin.Default()

    // Health check endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "OK"})
    })

    // Start the server
    r.Run(":8080") // Default port is 8080
}