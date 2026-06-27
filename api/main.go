package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/docplatform/api/handlers"
	"github.com/docplatform/api/middleware"
	"github.com/docplatform/api/services"
	_ "github.com/docplatform/shared/metrics" // register metrics on startup
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
	if err := services.InitTenantTable(); err != nil {
		slog.Error("Tenant table init failed", "error", err)
		os.Exit(1)
	}

	app := fiber.New(fiber.Config{
		AppName: "docplatform-api",
	})

	app.Use(middleware.Correlation) // Assign X-Request-ID before anything else

	app.Use(logger.New(logger.Config{
		Format: `{"time":"${time}","status":${status},"method":"${method}","path":"${path}","ip":"${ip}","request_id":"${locals:request_id}"}` + "\n",
	}))

	// Rate limiter: max 5 uploads per minute per IP. Only /upload is limited —
	// reads (/, /status, /events, /health, /metrics) are exempt so page loads and
	// the SSE→polling fallback (which hits /status every 2.5s) are never throttled.
	app.Use(limiter.New(limiter.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Path() != "/upload"
		},
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			slog.Warn("Rate limit hit", "ip", c.IP(), "request_id", c.Locals("request_id"))
			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded. Max 5 uploads per minute.",
			})
		},
	}))

	// Start SSE notification poller (requires InitSQS to have run)
	connManager := services.NewConnectionManager()
	go services.StartNotificationPoller(connManager)

	// Unauthenticated endpoints
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "api"})
	})
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Frontend is a static page with no secrets — served without auth so the
	// browser can load it and then collect an API key for the calls below.
	app.Get("/", handlers.Frontend)

	// All routes below are protected by API key authentication
	app.Use(middleware.Auth)

	app.Post("/upload", handlers.Upload)
	app.Get("/status/:file_id", handlers.Status)
	app.Get("/events/:file_id", func(c *fiber.Ctx) error {
		return handlers.Events(c, connManager)
	})
	app.Post("/admin/tenants", handlers.CreateTenant)

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
