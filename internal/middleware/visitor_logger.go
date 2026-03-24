package middleware

import (
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/kaantanis/urlshortener/internal/model"
)

const visitorKey = "visitorMeta"

func VisitorLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if ip == "" {
			remote := c.Context().RemoteIP()
			if remote != nil {
				ip = remote.String()
			}
		}
		if ip == "" || net.ParseIP(ip) == nil {
			ip = "unknown"
		}

		entry := model.VisitLog{
			IP:         ip,
			UserAgent:  c.Get("User-Agent"),
			Referer:    c.Get("Referer"),
			AcceptLang: c.Get("Accept-Language"),
			Origin:     c.Get("Origin"),
			Host:       c.Hostname(),
		}

		c.Locals(visitorKey, entry)
		return c.Next()
	}
}

func GetVisitorMeta(c *fiber.Ctx) model.VisitLog {
	value := c.Locals(visitorKey)
	if value == nil {
		return model.VisitLog{}
	}

	entry, ok := value.(model.VisitLog)
	if !ok {
		return model.VisitLog{}
	}
	return entry
}
