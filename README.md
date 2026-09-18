# LinkForge

A production-oriented URL shortening service focused on fast redirects, cache-aware persistence, and practical API controls.

## What it does

- Generates compact IDs using the Snowflake algorithm with Base62 encoding.
- Uses Redis as a cache-aside layer for fast redirect lookups.
- Persists links and metadata in PostgreSQL (Supabase).
- Supports custom aliases with validation and conflict handling.
- Supports configurable link expiration from 5 minutes to 30 days.
- Tracks click counts and the latest redirect timestamp.
- Applies Redis-backed rate limiting to protect write/read endpoints.
- Validates request content types, payload size, and URL format.
- Includes a React + TypeScript web client with local link history.

## Architecture

```text
Browser
  │
  ▼
React + TypeScript
  │ POST /api/shorten
  ▼
Go + chi HTTP API
  ├── Redis ── cache + rate limiting
  └── PostgreSQL ── durable link + analytics metadata

Redirect path:
Browser → GET /{id} → Redis fast path → PostgreSQL metadata → redirect
```

## API

### Create a short link

`POST /api/shorten`

```json
{
  "url": "https://example.com/a/very/long/path",
  "alias": "docs",
  "expires_in": 86400
}
```

`alias` and `expires_in` are optional. `expires_in` is expressed in seconds; valid non-zero values range from 5 minutes to 30 days.

### Link statistics

`GET /api/links/{id}/stats`

Returns the original URL, click count, last click timestamp, and expiry metadata.

## Project structure

```text
.
├── backend/
│   ├── cmd/api/
│   ├── internal/
│   └── migrations/
├── frontend/
├── infra/
└── package.json
```

## Local setup

1. Copy `.env.example` to `.env` and configure PostgreSQL/Supabase and Redis.
2. Start optional local infrastructure with `npm run infra:up`.
3. Run the backend with `npm run backend:run`.
4. Run the frontend with `npm run frontend:dev`.

## Tech stack

- **Backend:** Go, chi
- **Frontend:** React, TypeScript, Vite, Tailwind CSS
- **Database:** PostgreSQL / Supabase
- **Cache & rate limiting:** Redis / Upstash
- **ID generation:** Snowflake + Base62
- **Infrastructure:** Docker
