// main.go
package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/chromedp/chromedp"
    "github.com/gin-gonic/gin"
)

type HTMLToImageRequest struct {
    HTML     string `json:"html" binding:"required"`
    Width    int    `json:"width,omitempty"`
    Height   int    `json:"height,omitempty"`
    Quality  int    `json:"quality,omitempty"`
}

type HTMLToImageResponse struct {
    Success bool   `json:"success"`
    Image   string `json:"image,omitempty"` // base64 encoded image
    Error   string `json:"error,omitempty"`
}

func main() {
    router := gin.Default()
    
    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })

    // HTML to image endpoint
    router.POST("/convert", convertHTMLToImageHandler)

    // Middleware para rate limiting
    router.Use(rateLimitMiddleware())

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}