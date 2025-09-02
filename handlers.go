// handlers.go
package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/chromedp/cdproto/page"
    "github.com/chromedp/chromedp"
    "github.com/gin-gonic/gin"
)

func convertHTMLToImageHandler(c *gin.Context) {
    var req HTMLToImageRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, HTMLToImageResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    // Validar parámetros
    if req.Width <= 0 {
        req.Width = 1200
    }
    if req.Height <= 0 {
        req.Height = 800
    }
    if req.Quality <= 0 || req.Quality > 100 {
        req.Quality = 80
    }

    // Convertir HTML a imagen
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

func convertHTMLToImage(req HTMLToImageRequest) ([]byte, error) {
    // Configurar contexto con timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Crear instancia de Chrome
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

    // Ejecutar las acciones de Chrome
    err := chromedp.Run(taskCtx,
        chromedp.Navigate("about:blank"),
        chromedp.Evaluate(fmt.Sprintf(`
            document.write(%q);
            document.close();
        `, req.HTML), nil),
        chromedp.WaitReady("body"),
        chromedp.ActionFunc(func(ctx context.Context) error {
            // Capturar screenshot
            var err error
            buf, err = page.CaptureScreenshot().
                WithQuality(int64(req.Quality)).
                WithClip(&page.Viewport{
                    X:      0,
                    Y:      0,
                    Width:  float64(req.Width),
                    Height: float64(req.Height),
                    Scale:  1,
                }).Do(ctx)
            return err
        }),
    )

    return buf, err
}