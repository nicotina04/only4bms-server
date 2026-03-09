# Only4BMS Multiplayer Server

Lightweight self-hosted multiplayer server for the BMS rhythm game Only4BMS.

## Stack
- Go + gorilla/websocket + SQLite
- Target: Oracle Cloud Free Tier (ARM 4-core / 24GB)

## Project Structure
```
main.go              — Entrypoint, starts HTTP/WS server
config.go            — Env vars / CLI flags configuration
internal/
  ws/                — WebSocket hub, client, message types
  lobby/             — Lobby state machine, manager
  api/               — HTTP handlers (songs, daily)
  daily/             — Daily course seed, ranking (Phase 2)
  db/                — SQLite init, queries (Phase 2)
  auth/              — Discord OAuth2, JWT (Phase 4)
songs/               — Song data directory
documents/           — Design docs (not tracked by git)
```

## Design Documents
Detailed design specs live in `documents/` (git-ignored):
- `documents/BACKEND_DESIGN.md` — Full backend design (API, WS protocol, DB schema, implementation order)
- `documents/MULTIPLAYER_API.md` — Client-server API spec

**If you need information not covered in these documents, ask the user.**

## Conventions
- Go standard project layout (`internal/`)
- WebSocket messages: `{"type": "event_name", "data": {...}}` JSON envelope
- Tests: `*_test.go` files, run with `go test ./...`
- Git flow: `main` → `develop` → `feature/*`

## Build & Run
```bash
go build -o only4bms-server .
./only4bms-server
```

## Environment Variables
```
PORT=8080
SONGS_DIR=./songs
DB_PATH=./rankings.db
SERVER_PASSWORD=        # optional
DAILY_RESET_HOUR=0
```

## Implementation Phases
1. **Phase 1** (current): Core server — HTTP + WS + lobby + score relay
2. **Phase 2**: Daily course — SQLite + seed generation + ranking
3. **Phase 3**: Deployment — Docker + ARM cross-compile
4. **Phase 4**: Discord OAuth2 authentication
