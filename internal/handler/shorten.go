package handler

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/kaantanis/urlshortener/internal/service"
)

type ShortenHandler struct {
	urlService *service.URLService
}

func NewShortenHandler(urlService *service.URLService) *ShortenHandler {
	return &ShortenHandler{urlService: urlService}
}

type shortenRequest struct {
	URL string `json:"url"`
}

func (h *ShortenHandler) Handle(c *fiber.Ctx) error {
	var req shortenRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("invalid shorten request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid json body",
		})
	}

	created, shortURL, err := h.urlService.CreateShortURL(req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) || errors.Is(err, service.ErrURLTooLong) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		slog.Error("failed to create short url", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	slog.Info("short url created", "code", created.Code, "original", created.Original)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":       created.Code,
		"short_url":  shortURL,
		"original":   created.Original,
		"created_at": created.CreatedAt,
	})
}
