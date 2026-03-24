package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kaantanis/urlshortener/internal/db"
	"github.com/kaantanis/urlshortener/internal/handler"
	"github.com/kaantanis/urlshortener/internal/middleware"
	"github.com/kaantanis/urlshortener/internal/repository"
	"github.com/kaantanis/urlshortener/internal/service"
)

func TestShortenRedirectAndStatsPagination(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	code := createShortURL(t, app, "https://example.com/test-path")

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/"+code, nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("redirect request failed: %v", err)
		}
		if resp.StatusCode != 302 {
			t.Fatalf("unexpected redirect status: %d", resp.StatusCode)
		}
	}

	stats := waitForStats(t, app, code, 3, 3)
	if stats["hit_count"].(float64) != 3 {
		t.Fatalf("expected hit_count 3, got %v", stats["hit_count"])
	}

	page1 := getStats(t, app, code, 2, 0)
	page2 := getStats(t, app, code, 2, 2)

	p1 := page1["pagination"].(map[string]any)
	if p1["limit"].(float64) != 2 || p1["offset"].(float64) != 0 || p1["returned"].(float64) != 2 {
		t.Fatalf("unexpected page1 pagination: %+v", p1)
	}

	p2 := page2["pagination"].(map[string]any)
	if p2["limit"].(float64) != 2 || p2["offset"].(float64) != 2 || p2["returned"].(float64) != 1 {
		t.Fatalf("unexpected page2 pagination: %+v", p2)
	}
}

func setupTestApp(t *testing.T) (*fiber.App, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "urls.db")
	database, err := db.NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("init sqlite failed: %v", err)
	}

	urlRepo, err := repository.NewURLRepository(database)
	if err != nil {
		t.Fatalf("init url repository failed: %v", err)
	}
	logRepo, err := repository.NewVisitLogRepository(database)
	if err != nil {
		t.Fatalf("init visit log repository failed: %v", err)
	}

	svc := service.NewURLService(urlRepo, logRepo, "http://localhost:3000", 6, 1, 256)
	shortenHandler := handler.NewShortenHandler(svc)
	redirectHandler := handler.NewRedirectHandler(svc)
	statsHandler := handler.NewStatsHandler(svc)

	app := fiber.New(fiber.Config{ProxyHeader: fiber.HeaderXForwardedFor})
	app.Post("/api/shorten", shortenHandler.Handle)
	app.Get("/api/stats/:code", statsHandler.Handle)
	app.Get("/:code", middleware.VisitorLogger(), redirectHandler.Handle)

	cleanup := func() {
		svc.Close()
		_ = urlRepo.Close()
		_ = logRepo.Close()
		_ = database.Close()
	}

	return app, cleanup
}

func createShortURL(t *testing.T, app *fiber.App, target string) string {
	t.Helper()

	body := fmt.Sprintf(`{"url":"%s"}`, target)
	req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("shorten request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected shorten status: %d", resp.StatusCode)
	}

	var parsed map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode shorten response failed: %v", err)
	}

	code, ok := parsed["code"].(string)
	if !ok || code == "" {
		t.Fatalf("missing code in shorten response: %+v", parsed)
	}
	return code
}

func waitForStats(t *testing.T, app *fiber.App, code string, expectedHits int, expectedVisits int) map[string]any {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		stats := getStats(t, app, code, 20, 0)
		hitCount := int(stats["hit_count"].(float64))
		visits := stats["recent_visits"].([]any)
		if hitCount >= expectedHits && len(visits) >= expectedVisits {
			return stats
		}
		time.Sleep(50 * time.Millisecond)
	}

	return getStats(t, app, code, 20, 0)
}

func getStats(t *testing.T, app *fiber.App, code string, limit int, offset int) map[string]any {
	t.Helper()

	url := fmt.Sprintf("/api/stats/%s?limit=%d&offset=%d", code, limit, offset)
	req := httptest.NewRequest("GET", url, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("stats request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected stats status: %d", resp.StatusCode)
	}

	var parsed map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode stats response failed: %v", err)
	}
	return parsed
}
