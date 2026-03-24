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
- `TRUSTED_PROXIES`: Comma-separated proxy CIDRs/IPs allowed to provide forwarded headers
- `RATE_LIMIT_WINDOW_SEC`: Rate limiter window in seconds (default `60`)
- `SHORTEN_RATE_LIMIT_MAX`: Max requests per window for `POST /api/shorten` (default `60`)
- `REDIRECT_RATE_LIMIT_MAX`: Max requests per window for `GET /:code` (default `600`)
- `LOG_WORKERS`: Number of background workers for visit log inserts (default `2`)
- `LOG_QUEUE_SIZE`: Buffered queue size for async visit logs (default `2048`)

## Database Notes (SQLite)

- If the database file does not exist, the API creates it automatically on startup (and creates parent directories if needed).
- Keep SQLite for this API unless you observe sustained write contention or latency spikes.
- Current setup already enables WAL mode and adds an index for stats/log lookups.
- Additional SQLite tuning is enabled: `synchronous=NORMAL`, `busy_timeout=5000`, `temp_store=MEMORY`, `wal_autocheckpoint=1000`, `cache_size=-20000`.
- DB connection settings are tuned for SQLite write safety in WAL mode (`max_open_conns=1`, `max_idle_conns=1`).
- Consider migrating to PostgreSQL when one or more of these happen consistently:
  - Frequent `database is locked` errors under normal production traffic
  - Noticeable p95/p99 growth on redirect writes or stats reads
  - Need for multi-instance write scaling, replicas, or advanced operational tooling
- Runtime SQLite side files (`.db-wal`, `.db-shm`, `.db-journal`) are expected with WAL mode and are ignored by git.

## API Endpoints

## Public API Readiness

- Request IDs are added via middleware for traceability.
- Rate limiting is active on `POST /api/shorten` and `GET /:code`.
- Redirect logs are written through a buffered worker queue (non-blocking request path).
- Forwarded headers are trusted only when `TRUSTED_PROXIES` is configured.

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

Returns URL details, hit count, and paginated recent visits.

Query params:
- `limit` (optional): default `20`, max `100`
- `offset` (optional): default `0`

```bash
curl -s http://localhost:3000/api/stats/aB3xKp
```

```bash
curl -s "http://localhost:3000/api/stats/aB3xKp?limit=10&offset=20"
```

Example response:

```json
{
  "code": "aB3xKp",
  "original": "https://example.com/docs",
  "created_at": "2026-03-24 05:30:12",
  "hit_count": 42,
  "pagination": {
    "limit": 10,
    "offset": 20,
    "returned": 10
  },
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
