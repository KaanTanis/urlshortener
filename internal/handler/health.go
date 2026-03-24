package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	startedAt time.Time
}

func NewHealthHandler(startedAt time.Time) *HealthHandler {
	return &HealthHandler{startedAt: startedAt}
}

func (h *HealthHandler) Handle(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":         "ok",
		"uptime_seconds": time.Since(h.startedAt).Seconds(),
	})
}
