package main

import (
    "log/slog"
    "os"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/joho/godotenv"
    "mdtopdf/api/handlers"
    "mdtopdf/api/services"
)

func main() {
    // Load .env (ignored in production — env vars set by docker-compose)
    _ = godotenv.Load()

    // Initialize AWS clients
    if err := services.InitS3(); err != nil {
        slog.Error("S3 init failed", "error", err)
        os.Exit(1)
    }
    if err := services.InitDynamoDB(); err != nil {
        slog.Error("DynamoDB init failed", "error", err)
        os.Exit(1)
    }
    if err := services.InitSQS(); err != nil {
        slog.Error("SQS init failed", "error", err)
        os.Exit(1)
    }

    app := fiber.New(fiber.Config{
        AppName: "mdtopdf-api",
    })

    // Structured request logging
    app.Use(logger.New(logger.Config{
        Format: `{"time":"${time}","status":${status},"method":"${method}","path":"${path}","ip":"${ip}"}` + "\n",
    }))

    // Rate limiter: max 5 requests per minute per IP
    app.Use(limiter.New(limiter.Config{
        Max:        5,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            slog.Warn("Rate limit hit", "ip", c.IP())
            return c.Status(429).JSON(fiber.Map{
                "error": "Rate limit exceeded. Max 5 uploads per minute.",
            })
        },
    }))

    // Routes
    app.Post("/upload", handlers.Upload)
    app.Get("/status/:file_id", handlers.Status)
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok", "service": "api"})
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    slog.Info("API service starting", "port", port)
    if err := app.Listen(":" + port); err != nil {
        slog.Error("Server error", "error", err)
        os.Exit(1)
    }
}