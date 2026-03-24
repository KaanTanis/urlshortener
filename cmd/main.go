package main

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/kaantanis/urlshortener/internal/db"
	"github.com/kaantanis/urlshortener/internal/handler"
	"github.com/kaantanis/urlshortener/internal/middleware"
	"github.com/kaantanis/urlshortener/internal/repository"
	"github.com/kaantanis/urlshortener/internal/service"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := getEnv("PORT", "3000")
	baseURL := getEnv("BASE_URL", "http://localhost:"+port)
	dbPath := getEnv("DB_PATH", "./data/urls.db")
	codeLength := getIntEnv("CODE_LENGTH", 6)
	env := getEnv("ENV", "development")

	database, err := db.NewSQLite(dbPath)
	if err != nil {
		slog.Error("failed to init sqlite", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			slog.Error("failed to close db", "error", closeErr)
		}
	}()

	urlRepo, err := repository.NewURLRepository(database)
	if err != nil {
		slog.Error("failed to init url repository", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := urlRepo.Close(); closeErr != nil {
			slog.Error("failed to close url repository", "error", closeErr)
		}
	}()

	logRepo, err := repository.NewVisitLogRepository(database)
	if err != nil {
		slog.Error("failed to init visit log repository", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := logRepo.Close(); closeErr != nil {
			slog.Error("failed to close visit log repository", "error", closeErr)
		}
	}()

	urlService := service.NewURLService(urlRepo, logRepo, baseURL, codeLength)
	startTime := time.Now()

	shortenHandler := handler.NewShortenHandler(urlService)
	redirectHandler := handler.NewRedirectHandler(urlService)
	statsHandler := handler.NewStatsHandler(urlService)
	healthHandler := handler.NewHealthHandler(startTime)

	app := fiber.New(fiber.Config{
		AppName:     "urlshortener",
		ProxyHeader: fiber.HeaderXForwardedFor,
	})

	app.Get("/health", healthHandler.Handle)
	app.Post("/api/shorten", shortenHandler.Handle)
	app.Get("/api/stats/:code", statsHandler.Handle)
	app.Get("/:code", middleware.VisitorLogger(), redirectHandler.Handle)

	slog.Info("starting server", "port", port, "env", env, "base_url", baseURL)
	if err = app.Listen(":" + port); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
