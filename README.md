# URL Shortener API

Production-ready URL shortening API built with Go, Fiber, and SQLite.

## Features

- Create short codes for valid absolute HTTP/HTTPS URLs
- Redirect short codes with temporary `302` and no-cache headers
- Track visit metadata asynchronously for redirects
- Read URL stats including hit count and recent visits
- Structured logging with `log/slog`
- SQLite WAL mode and foreign key enforcement enabled

## Tech Stack

- Go `1.21+`
- `github.com/gofiber/fiber/v2`
- `github.com/mattn/go-sqlite3`
- `github.com/matoous/go-nanoid/v2`
- `github.com/joho/godotenv`

## Setup

1. Clone the repository
2. Copy env file:

   ```bash
   cp .env.example .env
   ```

3. Set environment values in `.env`
4. Install dependencies:

   ```bash
   go mod tidy
   ```

5. Run the API:

   ```bash
   go run ./cmd
   ```

Server runs on `PORT` (default `3000`).

## Environment Variables

- `PORT`: HTTP listen port (example: `3000`)
- `BASE_URL`: Base URL used to build `short_url` in shorten response
- `CODE_LENGTH`: NanoID code length for generated short codes (default `6`)
- `DB_PATH`: SQLite file path (example: `./data/urls.db`)
- `ENV`: Environment label for logging context (example: `development`)

## Database Notes (SQLite)

- Keep SQLite for this API unless you observe sustained write contention or latency spikes.
- Current setup already enables WAL mode and adds an index for stats/log lookups.
- Consider migrating to PostgreSQL when one or more of these happen consistently:
  - Frequent `database is locked` errors under normal production traffic
  - Noticeable p95/p99 growth on redirect writes or stats reads
  - Need for multi-instance write scaling, replicas, or advanced operational tooling

## API Endpoints

### `GET /health`

Returns service health and uptime.

```bash
curl -s http://localhost:3000/health
```

Example response:

```json
{
  "status": "ok",
  "uptime_seconds": 12.34567
}
```

### `POST /api/shorten`

Creates a short URL for a valid absolute HTTP/HTTPS URL.

```bash
curl -s -X POST http://localhost:3000/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/docs"}'
```

Example response:

```json
{
  "code": "aB3xKp",
  "short_url": "http://localhost:3000/aB3xKp",
  "original": "https://example.com/docs",
  "created_at": "2026-03-24 05:30:12"
}
```

### `GET /:code`

Redirects (`302`) to original URL.

```bash
curl -i http://localhost:3000/aB3xKp
```

Expected behavior:
- `302 Found` with `Location` header when code exists
- `404` JSON error when code does not exist
- Response includes `Cache-Control`, `Pragma`, and `Expires` headers to prevent redirect caching and ensure each open reaches the API

### `GET /api/stats/:code`

Returns URL details, hit count, and last 20 visits.

```bash
curl -s http://localhost:3000/api/stats/aB3xKp
```

Example response:

```json
{
  "code": "aB3xKp",
  "original": "https://example.com/docs",
  "created_at": "2026-03-24 05:30:12",
  "hit_count": 42,
  "recent_visits": [
    {
      "id": 88,
      "code": "aB3xKp",
      "visited_at": "2026-03-24 05:42:01",
      "ip": "127.0.0.1",
      "user_agent": "curl/8.7.1",
      "referer": "",
      "accept_lang": "",
      "origin": "",
      "host": "localhost"
    }
  ]
}
```
