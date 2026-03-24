package handler

import (
	"log/slog"
	"strconv"

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
	limit, offset := parsePagination(c)

	row, visits, err := h.urlService.GetStats(code, limit, offset)
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
		"code":       row.Code,
		"original":   row.Original,
		"created_at": row.CreatedAt,
		"hit_count":  row.HitCount,
		"pagination": fiber.Map{
			"limit":    limit,
			"offset":   offset,
			"returned": len(visits),
		},
		"recent_visits": visits,
	})
}

func parsePagination(c *fiber.Ctx) (int, int) {
	const (
		defaultLimit = 20
		maxLimit     = 100
		defaultOff   = 0
	)

	limit := defaultLimit
	if rawLimit := c.Query("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
			if parsed > maxLimit {
				limit = maxLimit
			} else {
				limit = parsed
			}
		}
	}

	offset := defaultOff
	if rawOffset := c.Query("offset"); rawOffset != "" {
		if parsed, err := strconv.Atoi(rawOffset); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
