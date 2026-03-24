package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/kaantanis/urlshortener/internal/middleware"
	"github.com/kaantanis/urlshortener/internal/service"
)

type RedirectHandler struct {
	urlService *service.URLService
}

func NewRedirectHandler(urlService *service.URLService) *RedirectHandler {
	return &RedirectHandler{urlService: urlService}
}

func (h *RedirectHandler) Handle(c *fiber.Ctx) error {
	code := c.Params("code")

	row, err := h.urlService.ResolveByCode(code)
	if err != nil {
		if service.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "not found",
			})
		}
		slog.Error("resolve code failed", "code", code, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	if err = h.urlService.IncrementHitCount(code); err != nil {
		slog.Error("increment hit count failed", "code", code, "error", err)
	}

	// Prevent permanent redirect caching so each open reaches the server and gets logged.
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	if redirErr := c.Redirect(row.Original, fiber.StatusFound); redirErr != nil {
		slog.Error("redirect failed", "code", code, "error", redirErr)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	visitor := middleware.GetVisitorMeta(c)
	visitor.Code = code

	h.urlService.LogVisitAsync(visitor)
	slog.Info("redirect served", "code", code, "to", row.Original)
	return nil
}
