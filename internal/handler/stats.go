package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/kaantanis/urlshortener/internal/service"
)

type StatsHandler struct {
	urlService *service.URLService
}

func NewStatsHandler(urlService *service.URLService) *StatsHandler {
	return &StatsHandler{urlService: urlService}
}

func (h *StatsHandler) Handle(c *fiber.Ctx) error {
	code := c.Params("code")

	row, visits, err := h.urlService.GetStats(code)
	if err != nil {
		if service.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "not found",
			})
		}
		slog.Error("get stats failed", "code", code, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":          row.Code,
		"original":      row.Original,
		"created_at":    row.CreatedAt,
		"hit_count":     row.HitCount,
		"recent_visits": visits,
	})
}
