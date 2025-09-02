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

// Estructuras de datos
type HTMLToImageRequest struct {
    HTML     string `json:"html" binding:"required"`
    Width    int    `json:"width,omitempty"`
    Height   int    `json:"height,omitempty"`
    Quality  int    `json:"quality,omitempty"`
}

type HTMLToImageResponse struct {
    Success bool   `json:"success"`
    Image   string `json:"image,omitempty"`
    Error   string `json:"error,omitempty"`
}

func main() {
    router := gin.Default()
    
    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })

    // Endpoint principal
    router.POST("/convert", convertHTMLToImageHandler)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

// Handler principal
func convertHTMLToImageHandler(c *gin.Context) {
    var req HTMLToImageRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, HTMLToImageResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    // Valores por defecto
    if req.Width <= 0 {
        req.Width = 1200
    }
    if req.Height <= 0 {
        req.Height = 800
    }
    if req.Quality <= 0 || req.Quality > 100 {
        req.Quality = 80
    }

    // Conversión
    imageData, err := convertHTMLToImage(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, HTMLToImageResponse{
            Success: false,
            Error:   fmt.Sprintf("Conversion failed: %v", err),
        })
        return
    }

    c.JSON(http.StatusOK, HTMLToImageResponse{
        Success: true,
        Image:   base64.StdEncoding.EncodeToString(imageData),
    })
}

// Función de conversión CORREGIDA
func convertHTMLToImage(req HTMLToImageRequest) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Configurar Chrome headless con viewport
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-dev-shm-usage", true),
        chromedp.WindowSize(req.Width, req.Height), // Esta es la forma correcta
    )

    allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
    defer cancel()

    taskCtx, cancel := chromedp.NewContext(allocCtx)
    defer cancel()

    var buf []byte

    // Ejecutar en Chrome - versión corregida
    err := chromedp.Run(taskCtx,
        chromedp.Navigate("about:blank"),
        chromedp.Evaluate(fmt.Sprintf(`
            document.write(%q);
            document.close();
        `, req.HTML), nil),
        chromedp.WaitReady("body"),
        chromedp.Sleep(1*time.Second), // Esperar a que se renderice
        chromedp.CaptureScreenshot(&buf),
    )

    return buf, err
}

// Versión alternativa usando Evaluate para setear viewport
func convertHTMLToImageAlternative(req HTMLToImageRequest) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )

    allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
    defer cancel()

    taskCtx, cancel := chromedp.NewContext(allocCtx)
    defer cancel()

    var buf []byte

    err := chromedp.Run(taskCtx,
        chromedp.Navigate("about:blank"),
        // Setear viewport via JavaScript
        chromedp.Evaluate(fmt.Sprintf(`
            document.write(%q);
            document.close();
            window.resizeTo(%d, %d);
        `, req.HTML, req.Width, req.Height), nil),
        chromedp.WaitReady("body"),
        chromedp.Sleep(1*time.Second),
        chromedp.CaptureScreenshot(&buf),
    )

    return buf, err
}